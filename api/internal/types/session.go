package types

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

// 知识块结构
type KnowledgeChunk struct {
	ID      int64  `json:"id"`      //知识块ID
	Title   string `json:"title"`   //知识标题
	Content string `json:"content"` //知识内容
}

// 会话储存接口更新
type SessionStore interface {
	GetMessages(chatId string, limit int) ([]VectorMessage, error)      //保存单条消息
	SaveMessage(chatId, role, content string) error                     //保存单条消息
	SaveKnowledge(title, content string) error                          //保存知识库
	RetrieveKnowledge(query string, topK int) ([]KnowledgeChunk, error) //检索知识库
}
