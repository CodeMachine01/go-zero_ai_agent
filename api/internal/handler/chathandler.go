package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"GoAgent/api/internal/logic"
	"GoAgent/api/internal/svc"
	"GoAgent/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// Go面试官聊天SSE流式接口
func ChatHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//设置SSE响应头
		setSSEHeader(w)
		flusher, _ := w.(http.Flusher)
		//立即刷新头部
		flusher.Flush()

		//处理请求
		var req types.InterviewAPPChatReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		//创建取消上下文
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel() //确保资源释放

		l := logic.NewChatLogic(ctx, svcCtx)
		respChan, err := l.Chat(&req)
		if err != nil {
			//httpx.ErrorCtx(r.Context(), w, err)
			sendSSEError(w, flusher, err.Error())
			return
		}

		//处理流式响应
		for {
			select {
			case <-ctx.Done():
				return
			case resp, ok := <-respChan:
				if !ok {
					fmt.Fprint(w, "event: end\ndata:{}\n\n") //结束标记
					flusher.Flush()
					return
				}

				//handler加个内容处理，符合前端markdown格式
				safeContent := strings.ReplaceAll(resp.Content, "\n", "\\n")
				safeContent = strings.ReplaceAll(safeContent, "\r", "\\r")
				//直接输出内容，不加JSON包装
				fmt.Fprintf(w, "data:%s\n\n", safeContent)
				flusher.Flush()

				if resp.IsLast {
					return
				}
			}
		}
	}
}

// setSSEHeader设置服务器推送事件（SSE)的响应头
func setSSEHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Accel-Buffering", "no")
	w.Header().Set("Transfer-Encoding", "chunked")
}

func sendSSEError(w http.ResponseWriter, flusher http.Flusher, errMsg string) {
	_, fprintf := fmt.Fprintf(w, "event: error\ndata: {\"error\":\"%s\"}\n\n", errMsg)
	if fprintf != nil {
		return
	}
	flusher.Flush()
}
