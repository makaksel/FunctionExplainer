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

func (c *Client) Ask(code string) string {
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
		panic(err)
	}
	defer resp.Body.Close()

	// Парсим ответ
	body, _ := io.ReadAll(resp.Body)
	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		panic(err)
	}

	// Проверяем ошибки
	if chatResp.Type != "message" {
		fmt.Printf("Error type: '%s', AI response: '%s'\n", chatResp.Type, chatResp.Error.Message)
	}

	return chatResp.Content[0].Text
}

func (c *Client) ProcessAll(data []parser.ParsedFiles) files.NestedMap {
	ticker := time.NewTicker(time.Minute / 45) // 45 запросов в минуту
	defer ticker.Stop()

	var (
		wg sync.WaitGroup
		mu sync.Mutex
	)

	var results = make(files.NestedMap)

	for _, item := range data {
		<-ticker.C // Ждём разрешения

		wg.Add(1)
		go func(part parser.ParsedFiles) {
			defer wg.Done()
			res := c.Ask(part.Data)

			// Защищаем запись в общий слайс
			mu.Lock()

			if strings.Contains(res, "нет функций") {
				mu.Unlock()
				return
			}

			re := regexp.MustCompile(`\n{2,}`)
			formatedRes := re.ReplaceAllString(res, "\n") + "\n"

			if results[item.DiffFile] == nil {
				results[item.DiffFile] = make(map[string]string)
				results[item.DiffFile][item.SrcFile] = formatedRes
			} else {
				results[item.DiffFile][part.SrcFile] += formatedRes
			}
			mu.Unlock()
			fmt.Printf("Process file: %s; Source: %s\n", item.DiffFile, part.SrcFile)
		}(item)
	}

	wg.Wait() // Ждём завершения всех горутин

	return results
}
