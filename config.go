package main

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

// Config The structure of the configuration
type Config struct {
	APIAddress     string `json:"api_address"`
	TokenLimit     int    `json:"token_limit"`
	SystemRole     string `json:"system_role"`
	PollingTimeout int    `json:"polling_timeout"`
	BotToken       string `json:"bot_token"`

	// Selecting the method of obtaining updates: "Polling" or "Webhook"
	UpdateMethod string `json:"update_method"`

	// Webhuk parameters (used if Updatemethod = "Webhook")
	WebhookDomain string `json:"webhook_domain"`
	WebhookPort   string `json:"webhook_port"`
	CertFile      string `json:"cert_file"`
	KeyFile       string `json:"key_file"`

	// Selecting the operating mode with LM Studio:
	// "stream" – Streaming-mode,
	// "full" – Waiting for a ready answer.
	LMStudioMode string `json:"lm_studio_mode"`

	Language string `json:"language"`
}

var (
	config         Config
	configFileName = "config.json"
	configMutex    sync.Mutex
)

// Loading configuration
func loadConfig() error {
	configMutex.Lock()
	defer configMutex.Unlock()

	data, err := os.ReadFile(configFileName)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &config)
}

// Preservation of the configuration
func saveConfig() error {
	configMutex.Lock()
	defer configMutex.Unlock()

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configFileName, data, 0644)
}

// Initialization of configuration: if there is no file, we create with default values.
func initConfig() {
	if _, err := os.Stat(configFileName); os.IsNotExist(err) {
		config = Config{
			APIAddress:     "http://localhost:1234",
			TokenLimit:     2048,
			SystemRole:     "You are a helpful assistant.",
			PollingTimeout: 60,
			BotToken:       "YOUR_TELEGRAM_BOT_TOKEN",
			UpdateMethod:   "polling",
			WebhookDomain:  "mybot.domain.com",
			WebhookPort:    "443",
			CertFile:       "cert.pem",
			KeyFile:        "key.pem",
			LMStudioMode:   "full", // Values: "Stream" or "Full"
			Language:       "en",
		}

		if err := saveConfig(); err != nil {
			log.Fatalf("Configuration conservation error: %v", err)
		}
	} else {
		if err := loadConfig(); err != nil {
			log.Fatalf("Configuration loading error: %v", err)
		}
	}
}
