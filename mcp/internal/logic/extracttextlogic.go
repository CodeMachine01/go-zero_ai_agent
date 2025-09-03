package logic

import (
	"context"
	"fmt"
	"io"
	"os"

	"GoAgent/mcp/internal/svc"
	"GoAgent/mcp/internal/utils"
	"GoAgent/mcp/types/mcp"
	"github.com/zeromicro/go-zero/core/logx"
)

type ExtractTextLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewExtractTextLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExtractTextLogic {
	return &ExtractTextLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 流式上传PDF并返回解析文本
func (l *ExtractTextLogic) ExtractText(stream mcp.PdfProcessor_ExtractTextServer) error {
	//接收元数据
	firstChunk, err := stream.Recv()
	if err != nil {
		logx.Errorf("接收元数据失败%v", err)
		return err
	}

	meta := firstChunk.GetMetadata()
	if meta == nil {
		return stream.SendAndClose(&mcp.PdfResponse{Error: "缺少元数据"})
	}

	//验证文件类型
	if meta.MimeType != "application/pdf" {
		return stream.SendAndClose(&mcp.PdfResponse{
			Error: "仅支持PDF文件",
		})
	}

	//创建临时文件
	tmpFile, err := os.CreateTemp("", "pdf-*.pdf")
	if err != nil {
		logx.Errorf("创建临时文件失效：%v", err)
		return err
	}

	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	//写入首块数据
	if chunk := firstChunk.GetChunk(); chunk != nil {
		if _, err := tmpFile.Write(chunk); err != nil {
			logx.Errorf("写入临时文件失败：%v", err)
			return err
		}
	}

	//接收并写入后续数据块
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			logx.Errorf("接收数据失败%v", err)
			return err
		}
		if chunk := req.GetChunk(); chunk != nil {
			if _, err := tmpFile.Write(chunk); err != nil {
				logx.Errorf("写入临时文件失败%v", err)
				return err
			}
		}
	}
	//解析PDF
	content, err := extractPdfText(tmpFile.Name())
	if err != nil {
		logx.Errorf("PDF解析失败%v", err)
		return stream.SendAndClose(&mcp.PdfResponse{
			Error: "PDF解析失败：" + err.Error(),
		})
	}
	fmt.Println("消息解析完成，打包发送给api:", content)

	return stream.SendAndClose(&mcp.PdfResponse{
		Content: content,
	})
}

// extractPdfText 从文件路径提取文本
func extractPdfText(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	return utils.ExtractPDFText(file)
}
