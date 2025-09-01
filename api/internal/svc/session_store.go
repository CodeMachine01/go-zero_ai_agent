package svc

import (
	"GoAgent/api/internal/types"
	"github.com/sashabaranov/go-openai"
	"sync"
	"time"
)

// 内存会话存储实现
type MemorySessionStore struct {
	sessions     map[string]*types.ChatSession //存储所有对话 key=chatId
	lastAccessed map[string]time.Time          //记录会话最后访问时间
	lock         sync.RWMutex                  //读写锁，来保证并发安全
}

// 初始化空的对话存储
func NewMemorySessionStore() *MemorySessionStore {
	return &MemorySessionStore{
		sessions:     make(map[string]*types.ChatSession),
		lastAccessed: make(map[string]time.Time),
	}
}

func (m *MemorySessionStore) GetSession(chatId string) (*types.ChatSession, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	session, exists := m.sessions[chatId]
	if !exists {
		//创建对话，带系统消息
		return &types.ChatSession{
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "你是一个专业的Go语言面试官，负责评估候选人的Go语言能力。请提出有深度的问题并评估回答。",
				},
			},
		}, nil
	}

	//更新访问时间
	m.lastAccessed[chatId] = time.Now()
	return session, nil
}

// 保存对话，上下文截断和更新存储
func (m *MemorySessionStore) SaveSession(chatId string, session *types.ChatSession) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	//上下文截断（保留系统消息和最近5轮对话）
	if len(session.Messages) > 10 {
		newMessages := []openai.ChatCompletionMessage{session.Messages[0]} //保留系统消息
		start := len(session.Messages) - 5
		if start < 1 {
			start = 1
		}
		newMessages = append(newMessages, session.Messages[start:]...)
		session.Messages = newMessages
	}
	//保存会话并更新访问时间
	m.sessions[chatId] = session
	m.lastAccessed[chatId] = time.Now()
	return nil
}

// 清理过期对话（可定期调用）
func (m *MemorySessionStore) CleanupExpiredSessions(maxAge time.Duration) {
	m.lock.Lock()
	defer m.lock.Unlock()

	now := time.Now()
	for chatId, lastAccessed := range m.lastAccessed {
		if now.Sub(lastAccessed) > maxAge {
			delete(m.sessions, chatId)
			delete(m.lastAccessed, chatId)
		}
	}
}
