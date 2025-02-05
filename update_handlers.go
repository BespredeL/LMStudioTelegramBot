package main

import (
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"io"
	"net/http"
	"strings"
)

// Processing updates for the "Full" or "Stream" mode depending on the settings
func processUpdate(update tgbotapi.Update) {
	data, _ := json.Marshal(update)
	logger.Debugf("Update received: %s", data)
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID
	user := update.Message.From
	username := user.UserName
	if username == "" {
		username = strings.TrimSpace(user.FirstName + " " + user.LastName)
	}

	botUser := addOrUpdateUser(user.ID, username)
	if err := saveUsers(); err != nil {
		logger.Errorf("Error saving users: %v", err)
	}

	if !botUser.Allowed {
		logger.Debugf("Access denied: ID: %d, Username: %w", user.ID, username)
		deniedMsg := tgbotapi.NewMessage(chatID, t("Access denied."))
		_, _ = bot.Send(deniedMsg)
		return
	}

	userMessage := update.Message.Text

	// The command handler
	if update.Message.IsCommand() {
		commandHandler(update)
		return
	}

	// Depending on the operating mode of LM Studio, select the call function:
	if config.LMStudioMode == "stream" {
		updateConversationContextStream(chatID, "user", userMessage)
		conversation := buildConversationForRequest(chatID)

		// We send the action "prints ..."
		typing := tgbotapi.NewChatAction(chatID, tgbotapi.ChatTyping)
		_, _ = bot.Send(typing)

		response, err := callLMStudioStream(selectedModel, conversation, chatID)
		if err != nil {
			logger.Errorf("Error calling LM Studio: %v", err)
			errMsg := tgbotapi.NewMessage(chatID, t("Error generating response."))
			_, _ = bot.Send(errMsg)
			return
		}
		response = convertToTelegramFormat(response)
		updateConversationContextStream(chatID, "assistant", response)
	} else { // "full"
		updateConversationContext(chatID, "user", userMessage)
		ctxMutex.Lock()
		conversation := make([]LMMessage, len(contexts[chatID]))
		copy(conversation, contexts[chatID])
		ctxMutex.Unlock()

		// Send a message-indicator
		typingMsg := tgbotapi.NewMessage(chatID, t("Bot is typing..."))
		typingMsgID, _ := bot.Send(typingMsg)

		response, err := callLMStudio(selectedModel, conversation)
		if err != nil {
			logger.Errorf("Error calling LM Studio: %v", err)
			errMsg := tgbotapi.NewMessage(chatID, t("Error generating response."))
			_, _ = bot.Send(errMsg)
			return
		}

		// We delete the indicator and send the answer
		deleteTypingMsg := tgbotapi.NewDeleteMessage(chatID, typingMsgID.MessageID)
		_, _ = bot.Send(deleteTypingMsg)

		response = convertToTelegramFormat(response)
		updateConversationContext(chatID, "assistant", response)
		respMsg := tgbotapi.NewMessage(chatID, response)
		respMsg.ParseMode = tgParseMode
		_, _ = bot.Send(respMsg)

		logger.Debugf("Message in telegram: %s", response)
	}
}

// HTTP Handler for Webhook
func webhookHandler(w http.ResponseWriter, r *http.Request) {
	var update tgbotapi.Update
	body, err := io.ReadAll(r.Body)

	if err != nil {
		logger.Errorf("Error reading request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &update); err != nil {
		logger.Errorf("Error parsing update: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	processUpdate(update)
	w.WriteHeader(http.StatusOK)
}

// Launch of Webhook server
func startWebhookServer() {
	domain := strings.TrimPrefix(strings.TrimPrefix(config.WebhookDomain, "http://"), "https://")
	domain = strings.TrimLeft(domain, "/")

	webhookURL := fmt.Sprintf("https://%s/webhook", domain)
	if config.WebhookPort != "" {
		webhookURL = fmt.Sprintf("https://%s:%s/webhook", domain, config.WebhookPort)
	}

	logger.Info("Checking the current webhook in Telegram...")
	currentWebhook, err := bot.GetWebhookInfo()
	if err != nil {
		logger.Errorf("Error retrieving webhook information: %v", err)
	}

	if currentWebhook.URL == webhookURL {
		logger.Infof("The webhook is already set to the URL: %s", webhookURL)
	} else {
		logger.Infof("Set up a webhook on the URL: %s", webhookURL)

		wh, err := tgbotapi.NewWebhook(webhookURL)
		if err != nil {
			logger.Fatalf("Error creating webhook: %v", err)
		}

		if _, err := bot.Request(wh); err != nil {
			logger.Fatalf("Webhook installation error: %v", err)
		}
	}

	http.HandleFunc("/webhook", webhookHandler)

	addr := fmt.Sprintf(":%s", config.WebhookPort)
	logger.Infof("Launching a webhook server on %s...", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		logger.Fatalf("HTTP server error: %v", err)
	}
}

// Launch of Long Polling (for Polling-mode)
func startLongPolling(stopChan <-chan struct{}) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = config.PollingTimeout
	updates := bot.GetUpdatesChan(u)
	logger.Infof("Long polling mode started with timeout %d sec.", config.PollingTimeout)
	for {
		select {
		case <-stopChan:
			logger.Info("Stop long polling")
			return
		case update, ok := <-updates:
			if !ok {
				return
			}
			go processUpdate(update)
		}
	}
}

// Command handler
func commandHandler(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID

	if update.Message.IsCommand() {
		msg := tgbotapi.NewMessage(chatID, "")

		switch update.Message.Command() {
		case "start":
			msg.Text = t("Hello! I'm a Telegram bot that uses LM Studio.")
		case "clear":
			clearConversationContext(chatID)
			msg.Text = t("Chat history cleared.")
		default:
			msg.Text = t("I don't know that command")
		}

		if _, err := bot.Send(msg); err != nil {
			logger.Panic(err)
		}
	}
}
