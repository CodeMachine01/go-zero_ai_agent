package logic

import (
	"context"
	"errors"
	"github.com/sashabaranov/go-openai"
	"io"
	"strings"

	"GoAgent/api/internal/svc"
	"GoAgent/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChatLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Go面试官聊天SSE流式接口
func NewChatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChatLogic {
	return &ChatLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChatLogic) Chat(req *types.InterviewAPPChatReq) (<-chan *types.ChatResponse, error) {
	ch := make(chan *types.ChatResponse)

	go func() {
		defer close(ch)

		//获取或创建会话
		session, err := l.svcCtx.SessionStore.GetSession(req.ChatId)
		if err != nil {
			l.Logger.Errorf("获取会话失败：: %v", err)
			return
		}

		//添加用户消息到会话历史
		userMessage := openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: req.Message,
		}
		session.Messages = append(session.Messages, userMessage)

		//创建OpenAI请求
		request := openai.ChatCompletionRequest{
			Model:       l.svcCtx.Config.OpenAI.Model,
			Messages:    session.Messages, //使用会话历史
			Stream:      true,
			MaxTokens:   l.svcCtx.Config.OpenAI.MaxTokens,
			Temperature: l.svcCtx.Config.OpenAI.Temperature,
		}

		//创建流式响应
		stream, err := l.svcCtx.OpenAIClient.CreateChatCompletionStream(l.ctx, request)
		if err != nil {
			l.Logger.Error(err)
			return
		}
		defer stream.Close()

		//收集完整响应内容
		var fullResponse strings.Builder

		for {
			select {
			case <-l.ctx.Done(): //上下文取消
				return
			default:
				response, err := stream.Recv()
				if errors.Is(err, io.EOF) {
					//流结束后保存会话
					assistantMessge := openai.ChatCompletionMessage{
						Role:    openai.ChatMessageRoleAssistant,
						Content: fullResponse.String(), //累积完整响应
					}
					session.Messages = append(session.Messages, assistantMessge)
					//持久化对话
					if err := l.svcCtx.SessionStore.SaveSession(req.ChatId, session); err != nil {
						l.Logger.Errorf("保存会话失败：: %v", err)
					}
					//发送结束标记
					ch <- &types.ChatResponse{IsLast: true}
					return
				}
				if err != nil {
					l.Logger.Error(err)
					return
				}

				if len(response.Choices) > 0 {
					content := response.Choices[0].Delta.Content
					if content != "" {
						//收集完整响应
						fullResponse.WriteString(content)

						ch <- &types.ChatResponse{
							Content: content,
							IsLast:  false,
						}
					}
				}
			}
		}
	}()

	return ch, nil
}
