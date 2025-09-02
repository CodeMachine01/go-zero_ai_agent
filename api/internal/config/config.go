package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	OpenAI struct {
		//基础配置
		ApiKey         string `json:"apiKey"`         //API密钥（本地部署留空）
		BaseURL        string `json:"baseUrl"`        //API基础地址
		Model          string `json:"model"`          //模型名称
		EmbeddingModel string `json:"embeddingModel"` //嵌入模型名称

		//核心生成参数
		MaxTokens        int     `json:"maxTokens"`
		Temperature      float32 `json:"temperature"`      //温度参数（0-2 ,越高越随机）
		TopP             float32 `json:"topP"`             //核心采样（0-1,越高越多样）
		PresencePenalty  float32 `json:"presencePenalty"`  //存在惩罚（-2.0到2.0）
		FrequencyPenalty float32 `json:"frequencyPenalty"` //频率惩罚（-2.0到2.0）
		Seed             *int    `json:"seed"`             //随机种子（-1表示随机）
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
