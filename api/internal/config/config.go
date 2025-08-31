package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	OpenAI struct {
		ApiKey      string
		BaseURL     string
		Model       string
		MaxTokens   int
		Temperature float32
	}
}
