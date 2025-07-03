package logutil

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
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
			// Robust: always write mock.log to project root
			cwd, err := os.Getwd()
			if err != nil {
				panic(err)
			}
			// Look for the project root by walking up until go.mod is found
			root := cwd
			for {
				if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
					break
				}
				parent := filepath.Dir(root)
				if parent == root {
					panic("project root (go.mod) not found")
				}
				root = parent
			}
			logPath := filepath.Join(root, "mock.log")
			logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
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
