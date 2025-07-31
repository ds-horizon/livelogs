package main

import (
	"embed"

	"github.com/dream11/livelogs/cmd"
)

//go:embed scripts/*
var SetupScript embed.FS

func main() {
	cmd.SetSetupScript(SetupScript)
	cmd.Execute()
}
