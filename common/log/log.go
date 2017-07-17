package log

import (
	"io/ioutil"
	"log"
	"os"
)

var (
	Debug *log.Logger
	Info *log.Logger
	Warn *log.Logger
	Error *log.Logger
)

func init() {
	debugHandle := ioutil.Discard
	infoHandle := os.Stdout
	warnHandle := os.Stdout
	errorHandle := os.Stderr

	verbose := os.Getenv("VERBOSE") == "1"
	if verbose {
		debugHandle = os.Stdout
	}

	Debug = log.New(debugHandle, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile)
	Info = log.New(infoHandle, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile)
	Warn = log.New(warnHandle, "[WARN] ", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(errorHandle, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)
}