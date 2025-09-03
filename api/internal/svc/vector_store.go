package svc

import (
	"GoAgent/api/internal/config"
	"GoAgent/api/internal/types"
	"GoAgent/api/internal/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sashabaranov/go-openai"
	"time"
)

// 向量存储结构
type VectorStore struct {
	Pool           *pgxpool.Pool  //数据库连接池
	OpenAIClient   *openai.Client //OpenAI客户端
	EmbeddingModel string         //向量模型名称
}

// 初始化向量存储
func NewVectorStore(cfg config.VectorDBConfig, openaiClient *openai.Client) (*VectorStore, error) {
	//构建连接字符串
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	//解析配置
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}
	poolConfig.MaxConns = int32(cfg.MaxConn) //设置最大连接数

	//创建连接池
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, err
	}

	return &VectorStore{
		Pool:           pool,
		OpenAIClient:   openaiClient,
		EmbeddingModel: cfg.EmbeddingModel,
	}, nil
}

// 保存消息到向量数据库
func (vs *VectorStore) SaveMessage(chatId, role, content string) error {
	//生成文本向量
	embedding, err := vs.generateEmbedding(content)
	if err != nil {
		return fmt.Errorf("生成嵌入失败：%w", err)
	}

	//将向量转换为JSON格式
	embeddingJSON, err := json.Marshal(embedding)
	if err != nil {
		return fmt.Errorf("序列化嵌入失败：%w", err)
	}
	//插入数据库
	//sql := `INSERT INTO vector_store (chat_id,role,content,embedding) VALUES ($1,$2,$3,$4)`
	sql := `INSERT INTO vector_store (chat_id,role,content,embedding,source_type) VALUES ($1,$2,$3,$4,'message')`
	_, err = vs.Pool.Exec(context.Background(), sql, chatId, role, content, embeddingJSON)
	return err
}

// 知识库保存
func (vs *VectorStore) SaveKnowledge(title, content string, cfg config.VectorDBConfig) error {
	fmt.Println("进入保存处理！！！：", cfg.Knowledge.MaxChunkSize)
	//分块处理知识内容 todo
	chunks := utils.SplitText(content, cfg.Knowledge.MaxChunkSize)
	fmt.Println("分块处理内容！！：")
	for _, chunk := range chunks {
		fmt.Println("循环插入中！！：")
		embedding, err := vs.generateEmbedding(chunk)
		if err != nil {
			return fmt.Errorf("生成嵌入失败：%w", err)
		}

		embeddingJSON, err := json.Marshal(embedding)
		if err != nil {
			return fmt.Errorf("序列化嵌入失败：%w", err)
		}
		sql := `INSERT INTO knowledge_base (title,content,embedding) VALUES ($1,$2,$3)`
		_, err = vs.Pool.Exec(context.Background(), sql, title, chunk, embeddingJSON)
		if err != nil {
			return err
		}
		fmt.Println("插入成功！！：")
	}
	return nil
}

// 知识检索
func (vs *VectorStore) RetrieveKnowldge(query string, topK int) ([]types.KnowledgeChunk, error) {
	queryEmbedding, err := vs.generateEmbedding(query)
	if err != nil {
		return nil, fmt.Errorf("生成查询嵌入失败%w", err)
	}
	queryEmbeddingJSON, err := json.Marshal(queryEmbedding)
	if err != nil {
		return nil, fmt.Errorf("序列化查询嵌入失败%w", err)
	}

	//使用余弦相似度检索
	sql := `SELECT id,title,content FROM knowledge_base ORDER BY embedding::jsonb::text <-> $1::text LIMIT $2`
	rows, err := vs.Pool.Query(context.Background(), sql, queryEmbeddingJSON, topK)
	if err != nil {
		return nil, fmt.Errorf("知识检索失败：%w", err)
	}
	defer rows.Close()

	var results []types.KnowledgeChunk
	for rows.Next() {
		var id int64
		var title, content string
		if err := rows.Scan(&id, &title, &content); err != nil {
			return nil, fmt.Errorf("扫描结果失败：%w", err)
		}
		results = append(results, types.KnowledgeChunk{
			ID:      id,
			Title:   title,
			Content: content,
		})
	}
	return results, nil
}

// 获取会话历史消息
func (vs *VectorStore) GetMessages(chatId string, limit int) ([]types.VectorMessage, error) {
	//查询数据库
	sql := `SELECT role,content FROM vector_store WHERE chat_id=$1 ORDER BY created_at DESC LIMIT $2`

	rows, err := vs.Pool.Query(context.Background(), sql, chatId, limit)
	if err != nil {
		return nil, fmt.Errorf("数据库查询失败: %w", err)
	}
	defer rows.Close()

	//处理查询结果
	var messages []types.VectorMessage
	for rows.Next() { //逐行获取查询结果
		var role, content string
		if err := rows.Scan(&role, &content); err != nil {
			return nil, fmt.Errorf("行扫描失败：: %w", err)
		}
		messages = append(messages, types.VectorMessage{
			Role:    role,
			Content: content,
		})
	}

	//反转消息顺序（最新消息在最后）
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	return messages, nil
}

// 生成文本向量
func (vs *VectorStore) generateEmbedding(text string) ([]float32, error) {
	if text == "" {
		return make([]float32, 1536), nil
	}
	//调用OpenAi Embedding API
	resp, err := vs.OpenAIClient.CreateEmbeddings(context.Background(),
		openai.EmbeddingRequest{
			Input: []string{text},
			Model: openai.EmbeddingModel(vs.EmbeddingModel),
		})
	if err != nil {
		return nil, fmt.Errorf("OpenAI API报错: %w", err)
	}
	if len(resp.Data) == 0 {
		return nil, errors.New("未返回嵌入数据")
	}
	return resp.Data[0].Embedding, nil
}

// 测试数据库连接
func (vs *VectorStore) TestConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return vs.Pool.Ping(ctx)
}
