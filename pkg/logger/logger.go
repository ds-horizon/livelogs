package logger

import (
	"fmt"
	"os"

	"github.com/mitchellh/cli"
)

type Logger struct{}

var userInterface = &cli.PrefixedUi{
	AskPrefix:       "",
	AskSecretPrefix: "[ SECRET ] ",
	OutputPrefix:    "",
	InfoPrefix:      "",
	ErrorPrefix:     "[ ERROR ] ",
	WarnPrefix:      "[ WARNING ] ",
	Ui: &cli.BasicUi{
		Writer:      os.Stdout,
		ErrorWriter: os.Stdout,
		Reader:      os.Stdin,
	},
}

const (
	successColor    = "\033[1;32m%s\033[0m"
	warningColor    = "\033[1;33m%s\033[0m"
	errorColor      = "\033[1;31m%s\033[0m"
	italicEmphasize = "\033[3m\033[1m%s\033[0m"
)

var isDebugModeEnabled = false

// Info : informative messages
func (l *Logger) Info(message string) {
	userInterface.Info(message)
}

// Success : success messages
func (l *Logger) Success(message string) {
	userInterface.Output(fmt.Sprintf(successColor, message))
}

// Warn : warning messages
func (l *Logger) Warn(message string) {
	userInterface.Warn(fmt.Sprintf(warningColor, message))
}

// Error : error/fatal messages
func (l *Logger) Error(message string) {
	userInterface.Error(fmt.Sprintf(errorColor, message))
}

func (l *Logger) ErrorAndExit(message string) {
	userInterface.Error(fmt.Sprintf(errorColor, message))
	os.Exit(1)
}

// Output : generic messages
func (l *Logger) Output(message string) {
	userInterface.Output(message)
}

// ItalicEmphasize : generic messages
func (l *Logger) ItalicEmphasize(message string) {
	userInterface.Output(fmt.Sprintf(italicEmphasize, message))
}

// EnableDebugMode : enable debug mode
func (l *Logger) EnableDebugMode() {
	isDebugModeEnabled = true
	l.Debug("Debug mode is enabled...")
}

// Debug : debugging messages
func (l *Logger) Debug(message string) {
	if isDebugModeEnabled {
		userInterface.Output(fmt.Sprintf("[ DEBUG ] %s", message))
	}
}
