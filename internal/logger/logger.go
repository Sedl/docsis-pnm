package logger

import (
	"log"
	"os"
)

type Logger struct {
	Debug *log.Logger
	Warning *log.Logger
	Error *log.Logger
}

func NewLogger () Logger {
	flags := log.Ldate | log.Ltime | log.Lmicroseconds
	logger := Logger{
		log.New(os.Stdout, "debug: ", flags),
		log.New(os.Stdout, "warning: ", flags),
		log.New(os.Stderr, "error: ", flags),
	}
	return logger
}