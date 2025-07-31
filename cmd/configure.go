package cmd

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dream11/livelogs/constants"
	"github.com/dream11/livelogs/pkg/shell"
	"github.com/spf13/cobra"
)

var setupScript embed.FS

func SetSetupScript(fs embed.FS) {
	setupScript = fs
}

func init() {
	configureCmd.Flags().BoolP(constants.ArgumentVerbose, "v", false, "verbose logging")
	rootCmd.AddCommand(configureCmd)
}

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "To run livelogs configuration script",
	Long:  "To run livelogs configuration script",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), constants.GlobalLogsCommandTimeout)
		defer cancel()
		select {
		case <-configureCmdHandler(ctx, cmd):
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

func configureCmdHandler(ctx context.Context, cmd *cobra.Command) <-chan string {
	result := make(chan string)
	isVerboseLoggingEnabled, _ := cmd.Flags().GetBool(constants.ArgumentVerbose)
	if isVerboseLoggingEnabled {
		log.EnableDebugMode()
	}

	go func() {
		defer close(result)
		tempDir := os.TempDir()
		scriptPath := filepath.Join(tempDir, "livelogs_setup.sh")
		err := os.WriteFile(scriptPath, setupScriptContent(), 0755)
		if err != nil {
			log.ErrorAndExit(fmt.Sprintf("Failed to write script to temporary file: %v", err))
		}

		// Ensure the script is executable
		err = os.Chmod(scriptPath, 0755)
		if err != nil {
			log.ErrorAndExit(fmt.Sprintf("Failed to set executable permissions on script: %v", err))
		}
		var exitCode = shell.Exec(fmt.Sprintf("sh %s", scriptPath))
		if exitCode != 0 {
			log.ErrorAndExit("Configuration script execution failed")
		}

		log.Info("✅ Configuration script executed successfully.")

		select {
		case <-ctx.Done():
			return
		case result <- "Command executed successfully.":
		}
	}()
	return result
}

func setupScriptContent() []byte {
	content, err := setupScript.ReadFile(constants.LivelogsSetupScriptPath)
	if err != nil {
		log.ErrorAndExit(fmt.Sprintf("Failed to read embedded script: %v", err))
	}
	return content
}
