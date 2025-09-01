package types

import "github.com/sashabaranov/go-openai"

// 会话结构体
type ChatSession struct {
	Messages []openai.ChatCompletionMessage `json:"messages"`
}

// 会话储存接口
type SessionStore interface {
	GetSession(chatId string) (*ChatSession, error)
	SaveSession(chatId string, session *ChatSession) error
}
