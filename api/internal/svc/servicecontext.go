package svc

import (
	"GoAgent/api/internal/config"
	"GoAgent/api/internal/types"
	openai "github.com/sashabaranov/go-openai"
)

type ServiceContext struct {
	Config       config.Config
	OpenAIClient *openai.Client
	SessionStore types.SessionStore //新增会话存储
}

func NewServiceContext(c config.Config) *ServiceContext {
	conf := openai.DefaultConfig(c.OpenAI.ApiKey)
	conf.BaseURL = c.OpenAI.BaseURL

	return &ServiceContext{
		Config:       c,
		OpenAIClient: openai.NewClientWithConfig(conf),
		SessionStore: NewMemorySessionStore(), //新增内存会话存储
	}
}
