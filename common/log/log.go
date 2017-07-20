package log

import (
	"io/ioutil"
	"log"
	"os"
	"io"
)

var (
	debug *log.Logger
	info *log.Logger
	error *log.Logger
)

func setupLogger(debugHandle, infoHandle, errorHandle io.Writer) {
	debug = log.New(debugHandle, "[DEBUG] ", log.Ldate|log.Ltime)
	info = log.New(infoHandle, "[INFO] ", log.Ldate|log.Ltime)
	error = log.New(errorHandle, "[ERROR] ", log.Ldate|log.Ltime)
}

func Debug(v ...interface{}) {
	debug.Println(v...)
}

func Info(v ...interface{}) {
	info.Println(v...)
}

func Error(v ...interface{}) {
	error.Println(v...)
}

func init() {
	debugHandle := ioutil.Discard
	infoHandle := os.Stdout
	errorHandle := os.Stderr

	verbose := os.Getenv("VERBOSE") == "1"
	if verbose {
		debugHandle = os.Stdout
	}

	setupLogger(debugHandle, infoHandle, errorHandle)
}