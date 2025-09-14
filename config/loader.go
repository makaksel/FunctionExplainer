package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Token         string
	SystemPrompt  string
	Proxy         string
	ExcludedPaths []string
}

func Load() Config {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	claudeToken := os.Getenv("CLAUDE_API_KEY")
	if claudeToken == "" {
		panic("CLAUDE_API_KEY is not set")
	}

	systemPrompt := os.Getenv("SYS_PROMPT")
	if systemPrompt == "" {
		panic("SYS_PROMPT is not set")
	}

	proxy := os.Getenv("PROXY")
	if proxy == "" {
		fmt.Println("PROXY is not set")
	}

	pathsStr := os.Getenv("EXCLUDED_PATHS")
	var excludedPaths []string
	if pathsStr != "" {
		excludedPaths = strings.Split(pathsStr, ",")
	}

	return Config{Token: claudeToken, SystemPrompt: systemPrompt, Proxy: proxy, ExcludedPaths: excludedPaths}
}
