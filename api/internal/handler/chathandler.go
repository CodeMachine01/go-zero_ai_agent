package handler

import (
	"GoAgent/api/internal/utils"
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
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
		////立即刷新头部
		//flusher.Flush()		//前端改为post请求后删除刷新

		//处理请求
		var req types.InterviewAPPChatReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		//处理PDF文件（如果有）
		var pdfContent string
		if file, header, err := r.FormFile("file"); err == nil {
			defer file.Close()

			//验证文件类型
			if header.Header.Get("Content-Type") != "application/pdf" {
				http.Error(w, "仅支持PDF文件", http.StatusBadRequest)
				return
			}

			//使用UniPDF提取文本
			//if content, err := utils.ExtractPDFText(file); err == nil {
			//	pdfContent = content
			//} else {
			//	logx.Error("PDF提取失败：%v", err)
			//}

			if content, err := svcCtx.PdfClient.ExtractText(file, header.Filename); err == nil {
				pdfContent = content
			} else {
				logx.Errorf("PDF提取失败：%v", err)
			}
		}

		//拼接消息
		req.Message = utils.CombineMessages(req.Message, pdfContent)
		fmt.Println("req.Message++++++7777:", pdfContent)
		fmt.Println("req.Message++++++6666:", req.Message)

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
					//fmt.Fprint(w, "event: end\ndata:{}\n\n") //结束标记
					flusher.Flush()
					return
				}

				//handler加个内容处理，符合前端markdown格式
				safeContent := strings.ReplaceAll(resp.Content, "\n", "\\n")
				safeContent = strings.ReplaceAll(safeContent, "\r", "\\r")
				//直接输出内容，不加JSON包装
				//fmt.Fprintf(w, "data:%s\n\n", safeContent)
				_, err := fmt.Fprintf(w, "data: %s\n\n", safeContent)
				if err != nil {
					return
				}
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
