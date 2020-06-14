package main

import (
	"os"
	"github.com/Olling/slog"
	"github.com/Olling/Enrolld/api"
	"github.com/Olling/Enrolld/dataaccess"
	"github.com/Olling/Enrolld/dataaccess/config"
)

func main() {
	config.InitializeConfiguration("/etc/enrolld/enrolld.conf")

	dataaccess.LoadOverwrites()
	dataaccess.LoadScripts()

	slog.SetLogLevel(slog.Debug)

	scriptPathErr := dataaccess.CheckScriptPath(config.Configuration.EnrollmentScriptPath)
	if scriptPathErr != nil {
		slog.PrintFatal("EnrollmentScriptPath Problem - stopping")
		os.Exit(1)
	}

	dataaccess.InitializeAuthentication()
	api.SetupRouter()
}
