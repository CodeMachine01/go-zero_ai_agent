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
	VectorDB VectorDBConfig //向量数据库配置
}

// 向量数据库配置
type VectorDBConfig struct {
	Host           string
	Port           int
	DBName         string
	User           string
	Password       string
	Table          string
	MaxConn        int
	EmbeddingModel string
}
