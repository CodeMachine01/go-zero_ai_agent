package logic

import (
	"GoAgent/api/internal/svc"
	"GoAgent/api/internal/types"
	"GoAgent/api/internal/utils"
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
)

type KnowledgeUploadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewKnowledgeUploadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *KnowledgeUploadLogic {
	return &KnowledgeUploadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *KnowledgeUploadLogic) KnowledgeUpload(req *types.KnowledgeUploadReq) (*types.KnowledgeUploadResp, error) {
	fmt.Println("进入logic处理！！：")
	//分块处理知识内容
	chunks := utils.SplitText(req.Content, l.svcCtx.Config.VectorDB.Knowledge.MaxChunkSize)
	fmt.Println("准备分块！！：")
	//保存每个分块
	for _, chunk := range chunks {
		if err := l.svcCtx.VectorStore.SaveKnowledge(req.Title,
			chunk, l.svcCtx.Config.VectorDB); err != nil {
			logx.Errorf("保存知识失败：%v", err)
			return nil, err
		}
	}
	fmt.Println("分块保存结束！！：")
	return &types.KnowledgeUploadResp{
		Msg:    "知识上传成功",
		Chunks: len(chunks),
	}, nil
}
