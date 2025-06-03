package internal

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type Logger struct {
	*log.Logger
	file *os.File
}

func NewLogger() *Logger {
	err := os.MkdirAll("logs", 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating logs directory: %v\n", err)
		return nil
	}

	logFile, err := os.OpenFile(
		filepath.Join("logs", "codegen.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating debug log file: %v\n", err)
		return nil
	}

	logger := log.New(logFile, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	return &Logger{
		Logger: logger,
		file:   logFile,
	}
}

func (l *Logger) Close() {
	if l.file != nil {
		l.file.Close()
	}
}

func (l *Logger) Log(format string, v ...interface{}) {
	if l.Logger != nil {
		l.Logger.Printf(format, v...)
	}
}
