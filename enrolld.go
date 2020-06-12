package main

import (
	"os"
	"github.com/Olling/slog"
	"github.com/Olling/Enrolld/api"
	"github.com/Olling/Enrolld/auth"
	"github.com/Olling/Enrolld/data"
	"github.com/Olling/Enrolld/config"
)

func main() {
	config.InitializeConfiguration("/etc/enrolld/enrolld.conf")

	fileio.LoadOverwrites()
	fileio.LoadScripts()

	slog.SetLogLevel(slog.Debug)

	scriptPathErr := fileio.CheckScriptPath(config.Configuration.EnrollmentScriptPath)
	if scriptPathErr != nil {
		slog.PrintFatal("EnrollmentScriptPath Problem - stopping")
		os.Exit(1)
	}

	auth.Initialize()
	api.SetupRouter()
}
