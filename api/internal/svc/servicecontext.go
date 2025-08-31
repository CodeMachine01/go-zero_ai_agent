package svc

import (
	"GoAgent/api/internal/config"
	openai "github.com/sashabaranov/go-openai"
)

type ServiceContext struct {
	Config       config.Config
	OpenAIClient *openai.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	conf := openai.DefaultConfig(c.OpenAI.ApiKey)
	conf.BaseURL = c.OpenAI.BaseURL

	return &ServiceContext{
		Config:       c,
		OpenAIClient: openai.NewClientWithConfig(conf),
	}
}
