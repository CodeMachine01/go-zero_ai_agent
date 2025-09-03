package types

type KnowledgeUploadReq struct {
	Title   string `form:"title"`
	Content string `form:"content"` //知识内容(由后端提取，前端不用传)
}

type KnowledgeUploadResp struct {
	Msg    string `json:"msg"`
	Chunks int    `json:"chunks"` //保存的知识块数量
}
