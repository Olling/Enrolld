package main

import (
	"os"
	"github.com/Olling/slog"
	"github.com/Olling/Enrolld/api"
	"github.com/Olling/Enrolld/fileio"
	"github.com/Olling/Enrolld/config"
)

func main() {
	config.InitializeConfiguration("/etc/enrolld/enrolld.conf")
	fileio.LoadOverwrites()

	scriptPathErr := fileio.CheckScriptPath()
	if scriptPathErr != nil {
		slog.PrintFatal("ScriptPath Problem - stopping")
		os.Exit(1)
	}

	api.SetupRouter()
}
