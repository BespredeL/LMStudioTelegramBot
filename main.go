package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	logLevel    = "debug"
	logFile     = "app_log.json"
	tgParseMode = "Markdown"
)

var bot *tgbotapi.BotAPI
var selectedModel string

func main() {
	setupLogger()

	logger.Info("Launch of the program ...")
	initConfig()
	logger.Info("The configuration is loaded")

	// We load the localization before starting
	logger.Info("Loading localization...")
	loadTranslations(config.Language)
	logger.Info("Localization is loaded")

	logger.Info("Users download...")
	if err := loadUsers(); err != nil {
		logger.Errorf("User download error: %v", err)
	}
	logger.Info("Users are successfully loaded")

	var errTg error
	bot, errTg = tgbotapi.NewBotAPI(config.BotToken)
	if errTg != nil {
		logger.Errorf("Telegram bot creation error: %v", errTg)
	} else {
		logger.Infof("Authorized the bot: %s", bot.Self.UserName)
	}

	logger.Info("GUI launch...")
	startGUI()
}
