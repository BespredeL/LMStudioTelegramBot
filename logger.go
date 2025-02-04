package main

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"sort"
	"strings"
	"time"
)

type LogEntry struct {
	Time    time.Time `json:"time"`
	Level   string    `json:"level"`
	Message string    `json:"msg"`
}

var logger *logrus.Logger

func setupLogger() {
	logger = logrus.New()
	logger.Formatter = &logrus.JSONFormatter{}

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logger.Fatalf("Failed to open the log file: %v", err)
	}

	// Log to the console and file at the same time
	multiWriter := io.MultiWriter(os.Stdout, file)
	logger.SetOutput(multiWriter)

	// Setting up the level of logistics
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logger.Warnf("Invalid log level '%s', defaulting to INFO", logLevel)
		level = logrus.InfoLevel // Default level
	}
	logger.SetLevel(level)
}

// Read log file
func readLog() (string, error) {
	file, err := os.Open(logFile)
	if err != nil {
		return "", fmt.Errorf("failed to open log file: %v", err)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			logger.Errorf("Error close file: %v", err)
		}
	}(file)

	var logs []LogEntry
	decoder := json.NewDecoder(file)
	for {
		var entry LogEntry
		if err := decoder.Decode(&entry); err != nil {
			if err.Error() == "EOF" {
				break
			}
			return "", fmt.Errorf("failed to read log entry: %v", err)
		}
		logs = append(logs, entry)
	}

	sort.SliceStable(logs, func(i, j int) bool {
		return logs[i].Time.After(logs[j].Time)
	})

	var result string
	for _, log := range logs {
		result += fmt.Sprintf("[%s] [%s] %s\n", log.Time.Format("2006-01-02 15:04:05"), strings.ToUpper(log.Level), log.Message)
	}

	return result, nil
}
