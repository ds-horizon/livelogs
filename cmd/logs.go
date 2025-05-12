package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/dream11/livelogs/constants"
	"github.com/dream11/livelogs/models"
	"github.com/dream11/livelogs/pkg/encryption"
	"github.com/dream11/livelogs/pkg/logger"
	"github.com/dream11/livelogs/protobuf"
	"github.com/dream11/livelogs/util"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/crypto/ssh"
	"google.golang.org/protobuf/proto"
)

var log logger.Logger

func init() {
	logsCmd.Flags().StringP(constants.ArgumentEnv, "e", "", "* environment name (Mandatory)")
	logsCmd.Flags().StringP(constants.ArgumentServiceName, "s", "", "service_name")
	logsCmd.Flags().StringP(constants.ArgumentComponentName, "c", "", "component_name")
	logsCmd.Flags().StringP(constants.ArgumentComponentType, "", "application", "component_type")
	logsCmd.Flags().StringP(constants.ArgumentOrg, "o", "d11", "org name can be: [d11, d3, dp, hulk]")
	logsCmd.Flags().StringP(constants.ArgumentCloudProvider, "", "aws", "cloud_provider can be: [aws, gcp] (Default is aws)")
	logsCmd.Flags().StringP(constants.ArgumentAccount, "a", "", "account type [prod, load, stag] (Default is based on env name if env is prod or uat then account is prod)")
	logsCmd.Flags().StringP(constants.ArgumentStartTime, "", "", "Start time if you want to see historic logs (Give the time in IST, with this format \"2025-01-02 15:04:05\")")
	logsCmd.Flags().StringP(constants.ArgumentEndTime, "", "", "End time if you want to see historic logs and wanted to see limited logs upto this time (Give the time in IST, with this format \"2006-01-02 15:04:05\")")
	logsCmd.Flags().StringP(constants.ArgumentSince, "", "", "When you want to see last 10 minute logs or last 1 hour logs just pass here as 10m or 1h")
	logsCmd.Flags().StringP(constants.ArgumentLinuxOperation, "l", "", "Linux operation you want to perform on streaming logs example  --linux_operation 'grep \"error\" | grep -iv \"user\"'")
	logsCmd.Flags().BoolP(constants.ArgumentVerbose, "v", false, "verbose logging")
	logsCmd.Flags().StringP(constants.LogSearchConfig, "", "", "Log search config")
	logsCmd.Flags().StringP(constants.ArgumentShowTags, "", "", "Comma-separated list of ddtags to display. If not specified, all ddtags will be shown by default.")

	// To enable debug mode
	_ = logsCmd.Flags().MarkHidden(constants.ArgumentVerbose)
	// Used only for central livelogs agent
	_ = logsCmd.Flags().MarkHidden(constants.LogSearchConfig)

	logsCmd.MarkFlagsRequiredTogether(constants.ArgumentEnv)
	rootCmd.AddCommand(logsCmd)
}

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "To print your component logs",
	Long:  "To print your component logs",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), constants.GlobalLogsCommandTimeout)
		defer cancel()
		select {
		case <-logsCmdHandler(ctx, cmd, args):
			log.Debug("Operation completed successfully.")
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				log.ErrorAndExit(fmt.Sprintf("Operation timed out after %v minutes", constants.GlobalLogsCommandTimeout.Minutes()))
			} else {
				log.Error(fmt.Sprintf("Something went wrong. Error: %v", ctx.Err()))
			}
		}
	},
}

func logsCmdHandler(ctx context.Context, cmd *cobra.Command, args []string) <-chan string {
	result := make(chan string)

	go func() {
		isVerboseLoggingEnabled, _ := cmd.Flags().GetBool(constants.ArgumentVerbose)
		if isVerboseLoggingEnabled {
			log.EnableDebugMode()
		}

		logCmdArgs := parseArguments(cmd)
		log.Debug(fmt.Sprintf("Command arguments: %+v", logCmdArgs))

		if util.IsCloudMachine() {
			log.Debug("Identified as central live log agent host")
			log.Success(fmt.Sprintf("Reading logs for service_name: %s component_name: %s env: %s org: %s account: %s cloudProvider: %s", logCmdArgs.ServiceName, logCmdArgs.ComponentName, logCmdArgs.Env, logCmdArgs.Org, logCmdArgs.Account, logCmdArgs.CloudProvider))
			var logSearchConfig models.LogSearchConfig
			if logCmdArgs.LogSearchConfig == (models.LogSearchConfig{}) {
				log.Debug("Log search config is empty so fetching...")
				logSearchConfig = util.GetLogsSearchConfig(logCmdArgs.Env, logCmdArgs.Org, logCmdArgs.Account, logCmdArgs.CloudProvider, logCmdArgs.ServiceName, logCmdArgs.ComponentName, logCmdArgs.ComponentType)
				validateArguments(&logCmdArgs, &logSearchConfig)
			} else {
				log.Debug("Log search config is not empty so using it...")
				logSearchConfig = logCmdArgs.LogSearchConfig
			}
			readFromKafka(&logCmdArgs, &logSearchConfig)
		} else {
			log.Debug("Identified as local live log agent host")
			logSearchConfig := util.GetLogsSearchConfig(logCmdArgs.Env, logCmdArgs.Org, logCmdArgs.Account, logCmdArgs.CloudProvider, logCmdArgs.ServiceName, logCmdArgs.ComponentName, logCmdArgs.ComponentType)
			validateArguments(&logCmdArgs, &logSearchConfig)
			commandForCentralLivelogsAgent := getCommandForCentralLivelogsAgent(cmd, args, logCmdArgs.LinuxOperation, &logSearchConfig)
			readFromCentralLivelogsAgent(commandForCentralLivelogsAgent, logSearchConfig, &logCmdArgs)
		}
		select {
		case <-ctx.Done():
			return
		case result <- "Command executed successfully.":
		}
	}()
	return result
}

func parseArguments(cmd *cobra.Command) models.LogsCommandArgs {
	env, _ := cmd.Flags().GetString(constants.ArgumentEnv)
	account, _ := cmd.Flags().GetString(constants.ArgumentAccount)
	serviceName, _ := cmd.Flags().GetString(constants.ArgumentServiceName)
	componentName, _ := cmd.Flags().GetString(constants.ArgumentComponentName)
	org, _ := cmd.Flags().GetString(constants.ArgumentOrg)
	cloudProvider, _ := cmd.Flags().GetString(constants.ArgumentCloudProvider)
	startTime, _ := cmd.Flags().GetString(constants.ArgumentStartTime)
	endTime, _ := cmd.Flags().GetString(constants.ArgumentEndTime)
	since, _ := cmd.Flags().GetString(constants.ArgumentSince)
	linuxOperation, _ := cmd.Flags().GetString(constants.ArgumentLinuxOperation)
	logSearchConfigString, _ := cmd.Flags().GetString(constants.LogSearchConfig)
	showTags, _ := cmd.Flags().GetString(constants.ArgumentShowTags)
	var logSearchConfig = models.LogSearchConfig{}
	if logSearchConfigString != "" {
		err := json.Unmarshal([]byte(logSearchConfigString), &logSearchConfig)
		if err != nil {
			log.Debug(fmt.Sprintf("Error unmarshalling log search config: %s. Error: %v", logSearchConfigString, err))
		}
	}

	return models.LogsCommandArgs{
		Env:             env,
		Account:         account,
		ServiceName:     serviceName,
		ComponentName:   componentName,
		Org:             org,
		CloudProvider:   cloudProvider,
		StartTime:       startTime,
		EndTime:         endTime,
		Since:           since,
		LinuxOperation:  linuxOperation,
		LogSearchConfig: logSearchConfig,
		ShowTags:        showTags,
	}
}

func validateArguments(args *models.LogsCommandArgs, config *models.LogSearchConfig) {
	if args.Since != "" {
		sinceDuration, err := time.ParseDuration(args.Since)

		if err != nil {
			log.ErrorAndExit("Error in parsing since as duration: " + args.Since)
		}

		validateIfLogsAreAvailable(sinceDuration, config, args.ServiceName, args.ComponentName)
	}

	if args.StartTime != "" {
		startTimeDuration := util.GetUtcTimeDuration(args.StartTime)
		validateIfLogsAreAvailable(startTimeDuration, config, args.ServiceName, args.ComponentName)
	}

	if args.EndTime != "" {
		endTimeDuration := util.GetUtcTimeDuration(args.EndTime)
		validateIfLogsAreAvailable(endTimeDuration, config, args.ServiceName, args.ComponentName)
	}
}

func validateIfLogsAreAvailable(duration time.Duration, config *models.LogSearchConfig, serviceName, componentName string) {
	if duration < 0 {
		log.ErrorAndExit("Future timestamp is not allowed")
	}

	if duration > time.Duration(config.MaxRetentionMinutes)*time.Minute {
		log.ErrorAndExit(fmt.Sprintf("We store only past %v minutes data for livelogs for service: %s and component: %s, for more logs please use grafana %s", config.MaxRetentionMinutes, serviceName, componentName, config.LogSearchGrafanaUrl))
	}
}

func getCommandForCentralLivelogsAgent(cmd *cobra.Command, args []string, linuxOperation string, logSearchConfig *models.LogSearchConfig) string {
	flags := ""
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		flagValue := flag.Value.String()
		if flagValue != "" && flagValue != "false" && flag.Name != "linux_operation" {
			if strings.Contains(flagValue, " ") {
				flagValue = "\"" + flagValue + "\""
			}
			flags += "--" + flag.Name + " " + flagValue + " "
		}
	})
	command := constants.CentralLiveLogAgentName + " " + cmd.Use + " " + flags
	if len(args) > 0 {
		command += " " + args[0]
		for _, arg := range args[1:] {
			command += " " + arg
		}
	}

	jsonString, err := json.Marshal(logSearchConfig)
	if err != nil {
		log.ErrorAndExit("Error in marshalling log search config: " + err.Error())
	}

	command += " --" + constants.LogSearchConfig + " '" + string(jsonString) + "'"

	if len(linuxOperation) > 0 {
		command += " | " + linuxOperation
	}
	log.Debug("Central live log agent command: " + command)
	return command
}

func getBrokersIpFromDns(hostname string) []string {
	log.Debug("Resolving DNS for Kafka brokers from hostname: " + hostname)
	var brokers []string
	ips := util.GetIpsFromHost(hostname)
	for _, ip := range ips {
		brokers = append(brokers, fmt.Sprintf("%s:%s", ip, constants.KafkaBrokerPort))
	}

	log.Debug("Resolved Kafka brokers: " + strings.Join(brokers, ", "))
	return brokers
}

func printLogsOnTerminal(dtags []byte, serviceName, hostname, message string) {
	var text string
	if dtags == nil || string(dtags) == constants.EmptyJSON {
		text = fmt.Sprintf("%s\t%s\t%s", serviceName, hostname, message)
	} else {
		text = fmt.Sprintf("%s\t%s\t%s\t%s", serviceName, hostname, string(dtags), message)
	}
	if strings.Contains(strings.ToLower(message), "error") {
		log.Error(text)
	} else {
		log.Info(text)
	}
}

func processMessage(msg, serviceName, componentName string, isLowerEnv bool, dtags []byte, logsStruct models.VectorLogsStruct) {
	shouldPrint := !isLowerEnv || (serviceName == "" && componentName == "") ||
		(logsStruct.Service != "" && strings.EqualFold(logsStruct.Service, serviceName) &&
			(componentName == "" || (logsStruct.Service != "" && strings.EqualFold(logsStruct.ComponentName, componentName))))
	if shouldPrint {
		printLogsOnTerminal(dtags, logsStruct.Service, logsStruct.Hostname, msg)
	}
}

func loadSamaraConfig() *sarama.Config {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Version = sarama.V2_8_0_0
	return config
}

func topicExists(brokerAddresses []string, topicName string, config *sarama.Config) (bool, error) {
	log.Debug(fmt.Sprintf("Checking if topic %s exists in Kafka", topicName))
	adminClient, err := sarama.NewClusterAdmin(brokerAddresses, config)
	if err != nil {
		log.ErrorAndExit("Failed to create Kafka admin client. Error: " + err.Error())
	}
	defer adminClient.Close()

	topics, err := adminClient.ListTopics()
	if err != nil {
		log.ErrorAndExit("Failed to list topics. Error: " + err.Error())
	}

	var topicNames []string
	for key := range topics {
		topicNames = append(topicNames, key)
	}

	log.Debug(fmt.Sprintf("Topics in Kafka: %v", topicNames))
	_, exists := topics[topicName]
	return exists, nil
}

func readFromKafka(args *models.LogsCommandArgs, logSearchConfig *models.LogSearchConfig) {
	log.Debug("Reading logs from Kafka")

	brokers := getBrokersIpFromDns(logSearchConfig.KafkaBrokerHost)
	samaraConfig := loadSamaraConfig()

	topicExists, err := topicExists(brokers, logSearchConfig.Topic, samaraConfig)

	if err != nil {
		log.ErrorAndExit("Failed to check if topic exists in Kafka: " + err.Error())
	}

	if !topicExists {
		log.ErrorAndExit("env:" + args.Env + " service_name:" + args.ServiceName + " component_name:" + args.ComponentName + " is not onboarded on Log Central")
	} else {
		log.Debug(fmt.Sprintf("Topic: %s exists in Kafka", logSearchConfig.Topic))
	}

	var wg sync.WaitGroup
	var closeOnce sync.Once
	var closeMutex sync.Mutex
	stopChan := make(chan struct{})
	var terminationWG sync.WaitGroup

	terminationWG.Add(1)
	go func() {
		defer terminationWG.Done()
		<-stopChan
		closeMutex.Lock()
		closeOnce.Do(func() {
			close(stopChan)
			log.Debug("Termination signal received. Closing stopChan.")
		})
		closeMutex.Unlock()
	}()

	client, err := sarama.NewClient(brokers, samaraConfig)
	if err != nil {
		log.ErrorAndExit(fmt.Sprintf("Failed to create Kafka client: %v", err))
	}

	defer func() {
		_ = client.Close()
	}()

	consumer, err := sarama.NewConsumer(brokers, samaraConfig)
	if err != nil {
		log.ErrorAndExit(fmt.Sprintf("Failed to create Kafka consumer: %v", err))
	}

	defer func() {
		err := consumer.Close()
		if err != nil {
			log.ErrorAndExit(fmt.Sprintf("Failed to close Kafka consumer: %v", err))
		}
	}()

	partitionConsumers := make(map[string]sarama.PartitionConsumer)
	partitions, err := consumer.Partitions(logSearchConfig.Topic)
	if err != nil {
		log.ErrorAndExit(fmt.Sprintf("Failed to get partitions for topic: %v", err))
	}

	for _, partition := range partitions {
		var sinceOffsets int64
		var partitionConsumer sarama.PartitionConsumer

		if args.Since != "" {
			duration, err := time.ParseDuration(args.Since)
			if err != nil {
				log.ErrorAndExit("Error in parsing duration for since: " + args.Since)
			}

			resultTime := time.Now().Add(-duration).UnixNano() / int64(time.Millisecond)
			sinceOffsets, err = client.GetOffset(logSearchConfig.Topic, partition, resultTime)
			if err != nil {
				log.ErrorAndExit(fmt.Sprintf("Failed to fetch offset from timestamp: %v", err))
			}

			partitionConsumer, err = consumer.ConsumePartition(logSearchConfig.Topic, partition, sinceOffsets)
			if err != nil {
				log.ErrorAndExit(fmt.Sprintf("Failed to create partition consumer: %v", err))
			}
		} else if args.Since == "" && args.StartTime != "" {
			startEpochTime := util.GetEpochTimeFromTimestamp(args.StartTime)
			startOffsets, err := client.GetOffset(logSearchConfig.Topic, partition, startEpochTime)
			if err != nil {
				log.ErrorAndExit(fmt.Sprintf("Failed to fetch offset from timestamp: %v", err))
			}

			partitionConsumer, err = consumer.ConsumePartition(logSearchConfig.Topic, partition, startOffsets)
			if err != nil {
				log.ErrorAndExit(fmt.Sprintf("Failed to create partition consumer: %v", err))
			}
		} else {
			partitionConsumer, err = consumer.ConsumePartition(logSearchConfig.Topic, partition, sarama.OffsetNewest)
			if err != nil {
				log.ErrorAndExit(fmt.Sprintf("Failed to create partition consumer: %v", err))
			}
		}

		key := fmt.Sprintf("%s-%d", logSearchConfig.Topic, partition)
		partitionConsumers[key] = partitionConsumer

		wg.Add(1)

		var endOffsets int64

		if args.Since == "" && args.EndTime != "" {
			endEpochTime := util.GetEpochTimeFromTimestamp(args.EndTime)
			endOffsets, err = client.GetOffset(logSearchConfig.Topic, partition, endEpochTime)
			if err != nil {
				log.ErrorAndExit(fmt.Sprintf("Failed to fetch offset from timestamp: %v", err))
			}
		}

		showTagsArray := strings.Split(args.ShowTags, ",")

		go func(consumer sarama.PartitionConsumer, endOffsets int64) {
			defer wg.Done()
			for eachMessage := range consumer.Messages() {
				if args.EndTime != "" && eachMessage.Offset >= endOffsets {
					closeMutex.Lock()
					closeOnce.Do(func() {
						close(stopChan)
					})
					closeMutex.Unlock()
					return
				}

				var vectorLogs = &protobuf.VectorLogs{}
				if err := proto.Unmarshal(eachMessage.Value, vectorLogs); err != nil {
					log.Debug("Failed to decode message value. Error: " + err.Error())
					continue
				}

				if args.ShowTags != "" {
					for key := range vectorLogs.Ddtags {
						if !isDdTagAllowed(key, showTagsArray) {
							delete(vectorLogs.Ddtags, key)
						}
					}
				}

				ddtags, err := json.Marshal(vectorLogs.Ddtags)
				if err != nil {
					log.Debug("Failed to encode ddtags. Error: " + err.Error())
					continue
				}

				logsStruct := models.VectorLogsStruct{
					Message:       vectorLogs.Message,
					Hostname:      util.DereferenceString(vectorLogs.Hostname),
					Env:           vectorLogs.Env,
					ComponentName: vectorLogs.ComponentName,
					Service:       vectorLogs.ServiceName,
					Ddtags:        vectorLogs.Ddtags,
				}

				message := logsStruct.Message
				if reflect.TypeOf(message).String() == "string" {
					msg := message.(string)
					processMessage(msg, args.ServiceName, args.ComponentName, logSearchConfig.IsLowerEnv, ddtags, logsStruct)
				} else {
					msg, err := json.Marshal(message)
					if err != nil {
						log.Debug("Failed to decode message. Error: " + err.Error())
						continue
					}
					processMessage(string(msg), args.ServiceName, args.ComponentName, logSearchConfig.IsLowerEnv, ddtags, logsStruct)
				}
			}
			closeMutex.Lock()
			closeOnce.Do(func() {
				close(stopChan)
				log.Debug("Termination signal received. Closing stopChan.")
			})
			closeMutex.Unlock()

		}(partitionConsumer, endOffsets)
	}

	// Wait for the termination goroutine to complete
	terminationWG.Wait()

	for _, partitionConsumer := range partitionConsumers {
		err := partitionConsumer.Close()
		if err != nil {
			log.ErrorAndExit(fmt.Sprintf("Failed to close partition consumer: %v", err))
		}
	}
}

func isDdTagAllowed(tag string, allowedTags []string) bool {
	for _, s := range allowedTags {
		if s == tag {
			return true
		}
	}
	return false
}

func readFromCentralLivelogsAgent(command string, logSearchConfig models.LogSearchConfig, logsCommandArgs *models.LogsCommandArgs) {
	log.Debug("Reading logs from central livelogs agent")

	centralLiveLogsAgentIp := util.GetAnyRandomIpFromHost(logSearchConfig.LiveLogAgentHost)

	if centralLiveLogsAgentIp == "" {
		log.ErrorAndExit("Unable to resolve DNS of central livelogs agent host. Please connect to correct vpn")
	}

	log.Debug("Connecting to central livelogs agent IP: " + centralLiveLogsAgentIp)
	remoteAddr := fmt.Sprintf("%s:%d", centralLiveLogsAgentIp, logSearchConfig.LiveLogAgentSshPort)
	decryptedPem, err := encryption.Decrypt(logSearchConfig.LiveLogAgentSshPemKey)

	if err != nil {
		log.Debug("Failed to decrypt ssh key: " + err.Error())
		log.ErrorAndExit("Failed to connect to central livelogs agent host.")
	}

	log.Success("Connecting to central livelogs agent...")

	logSearchConfig.LiveLogAgentSshPemKey = decryptedPem

	sshConfig := &ssh.ClientConfig{
		User: logSearchConfig.LiveLogAgentSshUser,
		Auth: []ssh.AuthMethod{
			getPemAuth(logSearchConfig.LiveLogAgentSshPemKey),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         constants.CentralLiveLogAgentSshTimeout,
	}

	client, err := ssh.Dial("tcp", remoteAddr, sshConfig)
	if err != nil {
		log.ErrorAndExit("Failed to connect to central livelogs agent: " + err.Error())
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		log.ErrorAndExit("Failed to create tcp session with central livelogs agent: " + err.Error())
	}
	defer session.Close()

	stdoutPipe, err := session.StdoutPipe()
	if err != nil {
		log.ErrorAndExit("Failed to fetch data from central livelogs agent:" + err.Error())
	}

	liveLogsUser, err := os.Hostname()
	if err != nil {
		log.Debug("Error in fetching hostname: " + err.Error())
		liveLogsUser = constants.EnvLivelogsUser
	}
	envVars := fmt.Sprintf("%s=\"%s\"", constants.EnvLivelogsUser, liveLogsUser)
	command = fmt.Sprintf("env %s %s", envVars, command)
	log.Debug(fmt.Sprintf("User: [%s] is executing command: [%s] on central live log agent", liveLogsUser, command))
	err = session.Start(command)
	if err != nil {
		log.ErrorAndExit("Command execution on central livelogs agent failed: " + err.Error())
	}

	go func() {
		util.UserLogFunc(logsCommandArgs, logSearchConfig.Tenant)
	}()

	go func() {
		reader := bufio.NewReader(stdoutPipe)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					log.Error("Error reading from central livelogs agent: " + err.Error())
				}
				break
			}
			fmt.Print(line)
		}
	}()

	err = session.Wait()
	if err != nil {
		log.ErrorAndExit("Command execution on central livelogs agent failed: " + err.Error())
	}

}

func getPemAuth(key string) ssh.AuthMethod {
	signer, err := ssh.ParsePrivateKey([]byte(key))
	if err != nil {
		log.Debug("Fail to connect to central live log agent host. Unable to parse private key: " + err.Error())
		log.ErrorAndExit("Failed to connect to central live log agent host")
	}
	return ssh.PublicKeys(signer)
}
