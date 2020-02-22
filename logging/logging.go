package logging

import (
	"io"
	"log"
)

var (
	InfoLog  *log.Logger
	ErrorLog *log.Logger
)

func InitializeLogging(infologHandle io.Writer, errorlogHandle io.Writer) {
	InfoLog = log.New(infologHandle, "INFO: ", log.Lshortfile)
	ErrorLog = log.New(errorlogHandle, "ERROR: ", log.Lshortfile)
	InfoLog.Println("Logging Initialized")
}
