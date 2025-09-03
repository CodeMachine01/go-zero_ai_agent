package handler

import (
	"GoAgent/api/internal/logic"
	"GoAgent/api/internal/svc"
	"GoAgent/api/internal/types"
	"errors"
	"fmt"
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
)

func KnowledgeUploadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//设置SSE响应头
		setSSEHeader(w)
		fmt.Println("进入上传知识库！！！")
		//获取文件
		file, header, err := r.FormFile("file")
		if err != nil {
			httpx.Error(w, err)
			return
		}
		defer file.Close()

		//验证PDF
		if header.Header.Get("Content-Type") != "application/pdf" {
			httpx.Error(w, errors.New("仅支持PDF问价"))
			return
		}
		////提取文本
		//content, err := utils.ExtractPDFText(file)
		//if err != nil {
		//	httpx.Error(w, err)
		//	return
		//}
		content, err := svcCtx.PdfClient.ExtractText(file, header.Filename)
		if err != nil {
			httpx.Error(w, err)
		}
		//获取标题（使用文件名）
		title := header.Filename
		fmt.Println("标题：", title)
		//调用Logic保存知识
		l := logic.NewKnowledgeUploadLogic(r.Context(), svcCtx)
		resp, err := l.KnowledgeUpload(&types.KnowledgeUploadReq{
			Title:   title,
			Content: content,
		})
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
