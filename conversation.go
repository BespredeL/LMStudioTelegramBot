package main

import (
	"regexp"
	"strings"
	"sync"
)

var (
	// Dialogue contexts for each chat
	contexts = make(map[int64][]LMMessage)
	ctxMutex = sync.Mutex{}
)

// Building the history of the request, taking into account the restrictions of tokens
func buildConversationForRequest(chatID int64) []LMMessage {
	ctxMutex.Lock()
	defer ctxMutex.Unlock()

	allMsgs := contexts[chatID]
	var result []LMMessage
	tokenCount := 0
	for i := len(allMsgs) - 1; i >= 0; i-- {
		msgTokens := len(strings.Fields(allMsgs[i].Content))
		if tokenCount+msgTokens > config.TokenLimit {
			break
		}
		tokenCount += msgTokens
		result = append(result, allMsgs[i])
	}

	// We unfold the cut for the chronological order
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result
}

// Context updating for the "Full" mode
func updateConversationContext(chatID int64, role, content string) {
	ctxMutex.Lock()
	defer ctxMutex.Unlock()

	if _, ok := contexts[chatID]; !ok {
		contexts[chatID] = []LMMessage{
			{Role: "system", Content: config.SystemRole},
		}
	}

	contexts[chatID] = append(contexts[chatID], LMMessage{Role: role, Content: content})
	trimConversation(chatID)
}

// Context updating for Streaming-mode (with the possibility of reset)
func updateConversationContextStream(chatID int64, role, content string) {
	ctxMutex.Lock()
	defer ctxMutex.Unlock()

	if _, ok := contexts[chatID]; !ok {
		contexts[chatID] = []LMMessage{
			{Role: "system", Content: config.SystemRole},
		}
	}

	contexts[chatID] = append(contexts[chatID], LMMessage{Role: role, Content: content})
}

// Context clear
func clearConversationContext(chatID int64) {
	contexts[chatID] = []LMMessage{
		{Role: "system", Content: config.SystemRole},
	}
}

// Function of the sub-count "tokens" (example on the number of words)
func countTokens(messages []LMMessage) int {
	total := 0
	for _, m := range messages {
		total += len(strings.Fields(m.Content))
	}
	return total
}

// Removing the oldest messages if the tokens limits are exceeded
func trimConversation(chatID int64) {
	msgs := contexts[chatID]
	for countTokens(msgs) > config.TokenLimit && len(msgs) > 1 {
		// Leave a system message
		msgs = append(msgs[:1], msgs[2:]...)
	}
	contexts[chatID] = msgs
}

// Converting markdown to telegram format
func convertToTelegramFormat(text string) string {
	// We hide the contents of <think> tags under spoilers
	thinkRe := regexp.MustCompile(`(?s)<think>(.*?)</think>`)
	text = thinkRe.ReplaceAllStringFunc(text, func(match string) string {
		content := thinkRe.ReplaceAllString(match, "$1")
		return "```\n" + strings.TrimSpace(content) + "```"
	})

	// Markdown transformation into Telegram-compatible format
	text = strings.ReplaceAll(text, "**", "*") // Bold -> bold
	text = strings.ReplaceAll(text, "__", "_") // Emphasis -> italics
	text = strings.ReplaceAll(text, "~~", "~") // Emphasis -> italics

	// We remove the headlines of Markdown
	headingRe := regexp.MustCompile(`(?m)^#{1,6}\s*`)
	text = headingRe.ReplaceAllString(text, "")

	logger.Debugf("convertToTelegramFormat: %v", text)

	return text
}
