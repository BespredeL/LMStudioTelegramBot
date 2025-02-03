package main

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var bot *tgbotapi.BotAPI
var selectedModel string

func main() {
	log.Println("Launch of the program ...")
	initConfig()
	log.Println("The configuration is loaded")

	// We load the localization before starting
	loadTranslations(config.Language)

	if err := loadUsers(); err != nil {
		log.Printf("User download error: %v", err)
	}
	log.Println("Users are successfully loaded")

	var err error
	bot, err = tgbotapi.NewBotAPI(config.BotToken)
	if err != nil {
		log.Printf("Telegram bot creation error: %v", err)
	} else {
		log.Printf("Authorized the bot: %s", bot.Self.UserName)
	}

	log.Println("GUI launch...")
	startGUI()
	log.Println("GUI launched")
}
