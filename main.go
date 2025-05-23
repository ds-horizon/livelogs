package main

import (
	"encoding/json"
	"fmt"

	"github.com/dream11/livelogs/app"
	"github.com/dream11/livelogs/cmd"
	logger2 "github.com/dream11/livelogs/pkg/logger"
	"github.com/dream11/livelogs/pkg/request"
)

var logger logger2.Logger

const GITHUB_TAGS_URL = "https://api.github.com/repos/dream11/livelogs/tags"

func main() {
	cmd.Execute()
	latestVersion := getLatestVersion()
	if latestVersion != "" && !isLatestVersion(app.App.Version, latestVersion) {
		logger.Info(fmt.Sprintf("\nYou are using livelogs version %s; however, version %s is available", app.App.Version, latestVersion))
		logger.Info("Upgrade to the latest version via command 'brew install dream11/tools/livelogs'")
	}
}

func getLatestVersion() string {
	req := request.Request{
		Method: "GET",
		URL:    GITHUB_TAGS_URL,
	}
	res := req.Make()
	if res.Error != nil {
		logger.Debug("Error making http request to fetch latest version: " + res.Error.Error())
		return ""
	}
	if res.StatusCode != 200 {
		logger.Debug("Invalid status code while checking latest version of Livelogs: " + fmt.Sprint(res.StatusCode))
		return ""
	}
	var jsonResponse []map[string]interface{}
	err := json.Unmarshal(res.Body, &jsonResponse)
	if err != nil {
		logger.Debug("Unable to unmarshal latest version response : " + err.Error())
		return ""
	}

	// return the latest tag
	return jsonResponse[0]["name"].(string)
}

func isLatestVersion(currentVersion string, latestVersion string) bool {
	return currentVersion == latestVersion
}
