package svc

import (
	"GoAgent/api/internal/config"
	openai "github.com/sashabaranov/go-openai"
	"log"
)

type ServiceContext struct {
	Config       config.Config
	OpenAIClient *openai.Client
	VectorStore  *VectorStore
}

func NewServiceContext(c config.Config) *ServiceContext {
	//创建OpenAI客户端
	openaiConfig := openai.DefaultConfig(c.OpenAI.ApiKey)
	openaiConfig.BaseURL = c.OpenAI.BaseURL
	openAIClient := openai.NewClientWithConfig(openaiConfig)

	//初始化向量存储
	vectorStore, err := NewVectorStore(c.VectorDB, openAIClient)
	if err != nil {
		log.Fatalf("初始化向量数据库失败: %v", err)
	}
	//测试数据库连接
	if err := vectorStore.TestConnection(); err != nil {
		log.Fatalf("向量数据库连接失败: %v", err)
	} else {
		log.Println("向量数据库连接成功")
	}

	return &ServiceContext{
		Config:       c,
		OpenAIClient: openAIClient,
		VectorStore:  vectorStore,
	}
}
