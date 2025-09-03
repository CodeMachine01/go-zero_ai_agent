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
	PdfClient    *PdfClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	////创建OpenAI客户端
	//openaiConfig := openai.DefaultConfig(c.OpenAI.ApiKey)
	//openaiConfig.BaseURL = c.OpenAI.BaseURL
	//openAIClient := openai.NewClientWithConfig(openaiConfig)

	//ollama deepseek r1 7b客户端
	openaiConfig := openai.DefaultConfig("") //Ollama无需密钥
	openaiConfig.BaseURL = c.OpenAI.BaseURL  //指向本地Ollama
	openAIClient := openai.NewClientWithConfig(openaiConfig)

	//初始化向量存储
	vectorStore, err := NewVectorStore(c.VectorDB, openAIClient)
	if err != nil {
		log.Fatalf("初始化向量数据库失败: %v", err)
	}

	////设置UniPDF key
	//err = license.SetMeteredKey(c.UniPDFLicense)
	//if err != nil {
	//	fmt.Printf("设置UniPDF许可证失败：: %v\n", err)
	//	//如果没有授权，UniPDF会添加水印
	//}

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
		PdfClient:    NewPdfClient(c.MCP.Endpoint),
	}
}
