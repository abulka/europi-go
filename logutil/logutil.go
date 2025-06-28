package logutil

import (
	"fmt"
	"log"
	"os"
	"sync"
)

var teaMode bool
var logFile *os.File
var logger *log.Logger
var once sync.Once

func SetTeaMode(enabled bool) {
	teaMode = enabled
}

func Println(v ...interface{}) {
	if teaMode {
		once.Do(func() {
			var err error
			logFile, err = os.OpenFile("mock.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				panic(err)
			}
			logger = log.New(logFile, "", log.LstdFlags)
		})
		logger.Println(v...)
	} else {
		fmt.Println(v...)
	}
}

func Close() {
	if logFile != nil {
		logFile.Close()
	}
}
