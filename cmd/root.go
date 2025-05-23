package cmd

import (
	"github.com/dream11/livelogs/app"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "livelogs",
	Short:   "Check your service logs",
	Long:    `Livelogs is a simple tool to check your service logs for any environments`,
	Version: app.App.Version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.ErrorAndExit("Error with the command executed: " + err.Error())
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
