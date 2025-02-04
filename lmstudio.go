package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	apiTimeout = 15 * 60 * time.Second
)

type LMRequest struct {
	Model    string      `json:"model"`
	Messages []LMMessage `json:"messages"`
}

type LMRequestStream struct {
	Model    string      `json:"model"`
	Messages []LMMessage `json:"messages"`
	Stream   bool        `json:"stream"`
}

type LMMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type LMChoice struct {
	Index   int       `json:"index"`
	Message LMMessage `json:"message"`
}

type LMResponse struct {
	ID      string     `json:"id"`
	Object  string     `json:"object"`
	Created int        `json:"created"`
	Model   string     `json:"model"`
	Choices []LMChoice `json:"choices"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	SystemFingerprint string `json:"system_fingerprint"`
}

type LMResponseChunk struct {
	ID                string `json:"id"`
	Object            string `json:"object"`
	Created           int    `json:"created"`
	Model             string `json:"model"`
	SystemFingerprint string `json:"system_fingerprint"`
	Choices           []struct {
		Index int `json:"index"`
		Delta struct {
			Role    string `json:"role,omitempty"`
			Content string `json:"content,omitempty"`
		} `json:"delta"`
		Logprobs     interface{} `json:"logprobs"`
		FinishReason interface{} `json:"finish_reason"`
	} `json:"choices"`
}

type LMModelData struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	OwnedBy string `json:"owned_by"`
}

type LMModelsResponse struct {
	Data   []LMModelData `json:"data"`
	Object string        `json:"object"`
}

// Creating a request URL
func createURL(path string) string {
	return strings.TrimRight(config.APIAddress, "/") + path
}

// Getting a list of models from LM Studio
func fetchModels() ([]string, error) {
	url := createURL("/models")
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("request error: %v", err)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Errorf("Error closing response: %v", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid status code: %d", resp.StatusCode)
	}

	var modelResponse LMModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&modelResponse); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	var models []string
	for _, m := range modelResponse.Data {
		models = append(models, m.ID)
	}

	return models, nil
}

// Calling LM Studio (full answer)
func callLMStudio(model string, conversation []LMMessage) (string, error) {
	reqBody := LMRequest{
		Model:    model,
		Messages: conversation,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	url := createURL("/chat/completions")
	client := &http.Client{Timeout: apiTimeout}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(data))

	if err != nil {
		return "", fmt.Errorf("LM Studio request error: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Errorf("Error closing response: %v", err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var lmResp LMResponse
	if err := json.Unmarshal(body, &lmResp); err != nil {
		return "", fmt.Errorf("error parsing response: %v", err)
	}

	if len(lmResp.Choices) == 0 {
		return "", fmt.Errorf("no answers available")
	}

	return lmResp.Choices[0].Message.Content, nil
}

// Calling LM Studio in Streaming mode
func callLMStudioStream(model string, conversation []LMMessage, chatID int64) (string, error) {
	reqBody := LMRequestStream{
		Model:    model,
		Messages: conversation,
		Stream:   true,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	url := createURL("/chat/completions")
	client := &http.Client{Timeout: apiTimeout}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("LM Studio request error: %v", err)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Errorf("Error closing response: %v", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("invalid status code: %d", resp.StatusCode)
	}

	// We send the initial message that we will edit
	initialMsg := tgbotapi.NewMessage(chatID, "...")
	sentMsg, err := bot.Send(initialMsg)
	if err != nil {
		return "", fmt.Errorf("error sending message: %v", err)
	}

	var fullResponse string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "data: ") {
			line = strings.TrimPrefix(line, "data: ")
		}

		if line == "[DONE]" {
			break
		}

		var chunk LMResponseChunk
		if err := json.Unmarshal([]byte(line), &chunk); err != nil {
			logger.Errorf("Chunk parsing error: %v", err)
			continue
		}

		if len(chunk.Choices) > 0 {
			partial := chunk.Choices[0].Delta.Content
			fullResponse += partial
			edit := tgbotapi.NewEditMessageText(chatID, sentMsg.MessageID, fullResponse)
			edit.ParseMode = tgParseMode
			_, _ = bot.Request(edit)
		}
	}

	if err := scanner.Err(); err != nil {
		return fullResponse, err
	}

	return fullResponse, nil
}
