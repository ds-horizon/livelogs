package util

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/dream11/livelogs/constants"
	"github.com/dream11/livelogs/models"
	"github.com/dream11/livelogs/pkg/logger"
	"github.com/dream11/livelogs/pkg/request"
)

var log logger.Logger

func getLogsSearchConfig(env, org, account, cloudProvider, serviceName, componentName, componentType string) models.LogSearchConfig {
	queryMap := map[string]string{
		"serviceName":   serviceName,
		"componentName": componentName,
		"componentType": componentType,
		"env":           env,
		"org":           org,
		"account":       account,
		"cloudProvider": cloudProvider,
	}
	log.Debug(fmt.Sprintf("Fetching logs search config with query: %v", queryMap))

	req := request.Request{
		Method: "GET",
		URL:    constants.CentralLiveLogAgentHost + "/live-logs-config",
		Query:  queryMap,
	}
	res := req.Make()
	if res.Error != nil {
		log.Debug("Error making http request to fetch log search config" + res.Error.Error())
		log.ErrorAndExit("Error in fetching log search config. Please connect to vpn and try again.")
	}

	if res.StatusCode != 200 {
		var errorBody struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		err := json.Unmarshal(res.Body, &errorBody)
		if err != nil {
			log.ErrorAndExit("Error in fetching log search config: " + string(res.Body))
		}
		log.ErrorAndExit("Error in fetching log search config: " + errorBody.Error.Message)
	}

	var responseBody struct {
		Data models.LogSearchConfig `json:"data"`
	}

	err := json.Unmarshal(res.Body, &responseBody)
	if err != nil {
		log.ErrorAndExit("Error in fetching log search config: " + err.Error())
	}

	log.Debug(fmt.Sprintf("Fetched logs search config: %v", responseBody.Data))
	return responseBody.Data
}

func GetUtcTimeDuration(timeStamp string) time.Duration {
	const timestampLayout = "2006-01-02 15:04:05"
	parsedTime, err := time.Parse(timestampLayout, timeStamp)
	if err != nil {
		log.ErrorAndExit(fmt.Sprintf("Error parsing timestamp: %s", timeStamp))
	}
	return time.Since(parsedTime.Add(-time.Minute * 330))
}

func GetEpochTimeFromTimestamp(timeStamp string) int64 {
	timestampLayout := "2006-01-02 15:04:05"
	parsedTime, err := time.Parse(timestampLayout, timeStamp)
	if err != nil {
		log.ErrorAndExit(fmt.Sprintf("Error parsing timestamp: %s", timeStamp))
	}
	return parsedTime.Add(-time.Minute*330).UnixNano() / int64(time.Millisecond)
}

func GetLogsSearchConfig(env, org, account, cloudProvider, serviceName, componentName, componentType string) models.LogSearchConfig {
	if account == "" && (env == "prod" || org == "uat") {
		account = "prod"
	}
	return getLogsSearchConfig(env, org, account, cloudProvider, serviceName, componentName, componentType)
}

func GetIpsFromHost(host string) []string {
	log.Debug("Resolving DNS of logs-agent host: " + host)
	command := exec.Command("dig", host, "+short")
	var out bytes.Buffer
	command.Stdout = &out

	err := command.Run()
	if err != nil {
		log.ErrorAndExit("Error in connecting to host: " + host + " Error:" + err.Error())
	}

	scanner := bufio.NewScanner(&out)
	var ips []string
	for scanner.Scan() {
		ip := scanner.Text()
		if isValidIP(ip) {
			ips = append(ips, ip)
		}
	}
	return ips
}

func GetAnyRandomIpFromHost(host string) string {
	ips := GetIpsFromHost(host)
	selectedIP := ""
	if len(ips) > 0 {
		selectedIP = ips[rand.Intn(len(ips))]
	}
	return selectedIP
}

func isValidIP(ip string) bool {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return false
	}
	for _, part := range parts {
		if len(part) == 0 {
			return false
		}
		for _, digit := range part {
			if digit < '0' || digit > '9' {
				return false
			}
		}
		num := 0
		for _, digit := range part {
			num = num*10 + int(digit-'0')
		}
		if num < 0 || num > 255 {
			return false
		}
	}
	return true
}

func UserLogFunc(args *models.LogsCommandArgs, tenant string) {
	hostName := os.Getenv(constants.EnvLivelogsUser)
	if len(hostName) == 0 {
		hostName, _ = os.Hostname()
	}
	log.Debug("Command invoked by: " + hostName)

	command := models.UserLogCommandStruct{
		Org:           args.Org,
		Env:           args.Env,
		ServiceName:   args.ServiceName,
		ComponentName: args.ComponentName,
		Account:       args.Account,
		StartTime:     args.StartTime,
		EndTime:       args.EndTime,
		Since:         args.Since,
	}

	commandMarshal, err := json.Marshal(command)
	if err != nil {
		log.Debug("Error in marshalling user log command" + err.Error())
	}

	req := request.Request{
		Method: "POST",
		URL:    constants.CentralLiveLogAgentHost + "/livelogs-user-log",
		Header: map[string]string{
			"X-Tenant-Name": tenant,
			"Content-Type":  "application/json",
		},
		Body: models.UserLogStruct{
			Hostname: hostName,
			Command:  string(commandMarshal),
		},
	}

	log.Debug("Logging user session for command: " + string(commandMarshal))

	res := req.Make()
	if res.Error != nil {
		log.Debug("Fail to log user session" + res.Error.Error())
	}
	if res.StatusCode != 200 {
		log.Debug("Fail to log user session. Invalid response: " + string(res.Body))
	}
}

func isAwsMachine() bool {
	req := request.Request{
		Method:  "GET",
		URL:     constants.AwsMetadataUrl,
		Timeout: 2 * time.Second,
	}

	res := req.Make()
	if res.Error != nil {
		log.Debug("Error making http request to fetch aws metadata: " + res.Error.Error())
		return false
	}
	if res.StatusCode/100 != 2 && res.StatusCode/100 != 4 {
		log.Debug("Invalid status code while checking aws metadata: " + fmt.Sprint(res.StatusCode))
		log.Debug("Invalid aws metadata response: " + string(res.Body))
		return false
	}

	log.Debug("This is a aws machine")
	return true
}

func isGcpMachine() bool {
	req := request.Request{
		Method:  "GET",
		URL:     constants.GcpMetadataUrl,
		Timeout: 2 * time.Second,
		Header: map[string]string{
			"Metadata-Flavor": "Google",
		},
	}

	res := req.Make()
	if res.Error != nil {
		log.Debug("Error making http request to fetch gcp metadata: " + res.Error.Error())
		return false
	}
	if res.StatusCode != 200 {
		log.Debug("Invalid status code while checking gcp metadata: " + fmt.Sprint(res.StatusCode))
		return false
	}

	log.Debug("This is a gcp machine")
	return true
}

func DereferenceString(ptr *string) string {
	if ptr != nil {
		return *ptr
	}
	return ""
}

func IsCloudMachine() bool {
	return isAwsMachine() || isGcpMachine()
}
