package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/makaksel/FunctionExplainerGo/internal/files"
	"github.com/makaksel/FunctionExplainerGo/internal/parser"
)

type Client struct {
	Token        string
	SystemPrompt string
	HttpClient   *http.Client
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model       string    `json:"model"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float32   `json:"temperature"`
	Messages    []Message `json:"messages"`
	Stream      bool      `json:"stream"`
}
type ResponseMessage []struct {
	Text string `json:"text"`
}

type RespErr struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type ChatResponse struct {
	Content ResponseMessage `json:"content"`
	Message string          `json:"message"`
	Type    string          `json:"type"`
	Error   RespErr         `json:"error"`
}

func NewClient(token, systemPrompt, proxy string) *Client {
	client := &http.Client{}

	if proxy != "" {
		// Настройка прокси
		proxyURL, _ := url.Parse(proxy)
		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
		client = &http.Client{Transport: transport}
	}

	return &Client{
		Token:        token,
		SystemPrompt: systemPrompt,
		HttpClient:   client,
	}
}

func (c *Client) Ask(code string) (string, error) {
	// JSON body
	reqBody := ChatRequest{
		Model:       "claude-3-7-sonnet-20250219",
		MaxTokens:   2000,
		Temperature: 0.25,
		Stream:      false,
		Messages: []Message{
			{
				Role:    "assistant",
				Content: c.SystemPrompt,
			},
			{
				Role:    "user",
				Content: code,
			},
		},
	}
	data, _ := json.Marshal(reqBody)

	// Создание HTTP-запроса
	req, _ := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.Token)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	// Проверка на превышение лимита
	if resp.StatusCode == http.StatusTooManyRequests {
		return "", fmt.Errorf("rate_limit")
	}

	// Проверка на загруженность сервера
	if resp.StatusCode == 529 {
		return "", fmt.Errorf("server_limit")
	}

	body, _ := io.ReadAll(resp.Body)
	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("json error: %w", err)
	}

	// Проверяем ошибки
	if chatResp.Type != "message" {
		return "", fmt.Errorf("Error type: '%s', AI response: '%s'\n", chatResp.Type, chatResp.Error.Message)
	}

	if len(chatResp.Content) == 0 {
		return "", fmt.Errorf("empty response")
	}

	return chatResp.Content[0].Text, nil
}

func (c *Client) ProcessAll(data []parser.ParsedFiles) files.NestedMap {
	ticker := time.NewTicker(time.Minute / 50) // 50 запросов в минуту
	defer ticker.Stop()

	var (
		wg          sync.WaitGroup
		mu          sync.Mutex
		results     = make(files.NestedMap)
		rateLimitMu sync.Mutex
	)

	for _, part := range data {
		<-ticker.C // ограничение по частоте

		wg.Add(1)
		go func(part parser.ParsedFiles) {
			defer wg.Done()

			for {
				rateLimitMu.Lock() // ждем, если кто-то в паузе
				rateLimitMu.Unlock()

				res, err := c.Ask(part.Data)

				if err != nil {
					if strings.Contains(err.Error(), "rate_limit") || strings.Contains(err.Error(), "server_limit") {
						fmt.Println("Limit. Pause for a minute...")

						rateLimitMu.Lock()
						time.Sleep(time.Minute)
						rateLimitMu.Unlock()

						continue // повторяем запрос после паузы
					}
					fmt.Printf("Request error %s: %v\n", part.SrcFile, err)
					return
				}

				mu.Lock()
				if !strings.Contains(strings.ToLower(res), "нет функций") {

					re := regexp.MustCompile(`\n{2,}`)
					formatedRes := re.ReplaceAllString(res, "\n") + "\n"

					if results[part.DiffFile] == nil {
						results[part.DiffFile] = make(map[string]string)
					}

					results[part.DiffFile][part.SrcFile] += formatedRes
				}
				mu.Unlock()

				fmt.Printf("Process file: %s; Source: %s\n", part.DiffFile, part.SrcFile)
				break
			}

		}(part)
	}

	wg.Wait() // Ждём завершения всех горутин

	return results
}
