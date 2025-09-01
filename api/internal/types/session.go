package types

import "github.com/sashabaranov/go-openai"

//// 会话结构体
//type ChatSession struct {
//	Messages []openai.ChatCompletionMessage `json:"messages"`
//}
//
//// 会话储存接口
//type SessionStore interface {
//	GetSession(chatId string) (*ChatSession, error)
//	SaveSession(chatId string, session *ChatSession) error
//}

// 向量存储消息结构
type VectorMessage struct {
	Role    string `json:"role"`    //消息角色
	Content string `json:"content"` //消息内容
}

// 会话储存接口更新
type SessionStore interface {
	GetSession(chatId string) ([]openai.ChatCompletionMessage, error) //获取消息历史
	SaveMessage(chatId, role, content string) error                   //保存单条消息
}
