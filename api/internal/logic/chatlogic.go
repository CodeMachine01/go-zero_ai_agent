package logic

import (
	"GoAgent/api/internal/utils"
	"context"
	"errors"
	"fmt"
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
		//初始化状态管理
		StateManager := NewStateManager(l.svcCtx)

		//1.保存用户消息到向量数据库
		if err := l.svcCtx.VectorStore.SaveMessage(
			req.ChatId,
			openai.ChatMessageRoleUser,
			req.Message,
		); err != nil {
			l.Logger.Errorf("保存用户消息失败：%v", err)
			//不返回，继续处理对话
		}

		//2.获取当前状态（确保初始化)
		currentState, err := StateManager.GetOrInitState(req.ChatId)
		if err != nil {
			l.Logger.Errorf("获取状态失败: %v", err)
			currentState = types.StateStart
		}

		//3.知识检索（RAG核心）
		knowledge, err := l.svcCtx.VectorStore.RetrieveKnowldge(req.Message, 3)
		if err != nil {
			l.Logger.Errorf("知识检索失败：%v", err)
			knowledge = []types.KnowledgeChunk{} //确保不为nil
		}

		////3. 获取会话历史
		//messages, err := l.getSessionHistory(req.ChatId, knowledge)
		//if err != nil {
		//	l.Logger.Error("获取会话历史失败：%v", err)
		//	ch <- &types.ChatResponse{Content: "系统错误：无法获取对话历史", IsLast: true}
		//	return
		//}

		//4.构建系统消息（带状态）
		messages, err := l.buildMessagesWithState(req.ChatId, currentState, knowledge)
		if err != nil {
			l.Logger.Errorf("构建消息失败：%v", err)
			ch <- &types.ChatResponse{Content: "系统错误：无法构建对话", IsLast: true}
		}

		////5.创建OpenAI请求（使用本地部署的大模型）
		//request := openai.ChatCompletionRequest{
		//	Model:            l.svcCtx.Config.OpenAI.Model,
		//	Messages:         messages,
		//	Stream:           true,
		//	MaxTokens:        l.svcCtx.Config.OpenAI.MaxTokens,
		//	Temperature:      l.svcCtx.Config.OpenAI.Temperature,
		//	TopP:             l.svcCtx.Config.OpenAI.TopP,
		//	FrequencyPenalty: l.svcCtx.Config.OpenAI.FrequencyPenalty,
		//	PresencePenalty:  l.svcCtx.Config.OpenAI.PresencePenalty,
		//	Seed:             l.svcCtx.Config.OpenAI.Seed,
		//}

		//5.创建OpenAI请求（使用本地部署的大模型）
		request := openai.ChatCompletionRequest{
			Model:       l.svcCtx.Config.OpenAI.Model,
			Messages:    messages,
			Stream:      true,
			MaxTokens:   l.svcCtx.Config.OpenAI.MaxTokens,
			Temperature: l.svcCtx.Config.OpenAI.Temperature,
		}

		//6.创建流式响应
		stream, err := l.svcCtx.OpenAIClient.CreateChatCompletionStream(l.ctx, request)
		if err != nil {
			l.Logger.Error("创建聊天完成流失败：%v", err)
			ch <- &types.ChatResponse{Content: "系统错误：无法连接AI服务", IsLast: true}
			return
		}
		defer stream.Close()

		//5.处理流式响应
		var fullResponse strings.Builder

		for {
			select {
			case <-l.ctx.Done(): //上下文取消
				return
			default:
				response, err := stream.Recv()
				if errors.Is(err, io.EOF) { //流结束
					finalResponse := fullResponse.String()
					//保存AI回复
					if finalResponse != "" {
						if saveErr := l.svcCtx.VectorStore.SaveMessage(
							req.ChatId,
							openai.ChatMessageRoleAssistant,
							finalResponse,
						); saveErr != nil {
							l.Logger.Errorf("保存助手消息失败：%v", saveErr)
						}
						//流结束后处理状态更新
						newState, err := StateManager.EvaluateAndUpdateState(req.ChatId, finalResponse)
						if err != nil {
							l.Logger.Errorf("更新状态失败：%v", err)
						} else {
							l.Logger.Infof("状态更新：%s->%s", currentState, newState)
						}
					}
					//发送结束标记
					ch <- &types.ChatResponse{IsLast: true}
					return
				}
				if err != nil {
					l.Logger.Error("接收流数据失败：%v", err)
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

//// 获取会话历史
//func (l *ChatLogic) getSessionHistory(chatId string,
//	knowledge []types.KnowledgeChunk) ([]openai.ChatCompletionMessage, error) {
//	//获取最近的10条信息（约5轮对话）
//	vectorMessages, err := l.svcCtx.VectorStore.GetMessages(chatId, 10)
//	if err != nil {
//		return nil, err
//	}
//
//	//构建系统消息-注入知识
//	systemMessage := "你是一个专业的Go语言面试官，负责评估候选人的Go语言能力。请提出有深度的问题并评估回答。"
//
//	if len(knowledge) > 0 {
//		systemMessage += "\n\n相关知识："
//		for i, k := range knowledge {
//			//限制知识片段长度
//			truncatedContent := utils.TruncateText(k.Content, l.svcCtx.Config.VectorDB.Knowledge.MaxContextLength)
//			systemMessage += fmt.Sprintf("\n[知识片段%d]%s:%s", i+1, k.Title, truncatedContent)
//		}
//	}
//	fmt.Println("检索的数据", systemMessage)
//
//	//转换为OpenAI消息格式
//	messages := []openai.ChatCompletionMessage{
//		{
//			Role:    openai.ChatMessageRoleSystem,
//			Content: systemMessage,
//		},
//	}
//
//	//添加历史消息
//	for _, msg := range vectorMessages {
//		messages = append(messages, openai.ChatCompletionMessage{
//			Role:    msg.Role,
//			Content: msg.Content,
//		})
//	}
//	return messages, nil
//}

// 构建带状态的消息
func (l *ChatLogic) buildMessagesWithState(chatId, currentState string,
	knowledge []types.KnowledgeChunk) ([]openai.ChatCompletionMessage, error) {
	//构建状态特定的系统消息
	systemMessage := "你是一个专业的Go语言面试官，负责评估候选人的Go语言能力。"
	systemMessage += "\n\n当前状态：" + currentState

	switch currentState {
	case types.StateStart:
		systemMessage += "\n目标: 欢迎候选人并开始面试流程"
	case types.StateQuestion:
		systemMessage += "\n目标: 提出有深度的问题考察Go语言核心概念"
	case types.StateFollowUp:
		systemMessage += "\n目标: 基于候选人的回答进行追问，深入考察理解深度"
	case types.StateEvaluate:
		systemMessage += "\n目标: 全面评估候选人的技术能力"
	case types.StateEnd:
		systemMessage += "\n目标: 结束面试并提供反馈"
	}

	//注入知识库
	if len(knowledge) > 0 {
		systemMessage += "\n\n相关背景知识："
		for i, k := range knowledge {
			//限制知识片段长度
			truncatedContent := utils.TruncateText(k.Content, l.svcCtx.Config.VectorDB.Knowledge.MaxContextLength)
			systemMessage += fmt.Sprintf("\n[知识片段%d]%s:%s", i+1, k.Title, truncatedContent)
		}
	}
	//fmt.Println("检索的数据", systemMessage)

	//转换为OpenAI消息格式
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemMessage,
		},
	}

	//获取最近的10条的历史信息（约5轮对话）
	historyMessages, err := l.svcCtx.VectorStore.GetMessages(chatId, 10)
	if err != nil {
		return nil, err
	}

	//添加历史消息
	for _, msg := range historyMessages {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}
	return messages, nil
}
