package svc

import (
	"GoAgent/mcp/types/mcp"
	"context"
	"errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"io/ioutil"
	"mime/multipart"
)

type PdfClient struct {
	client mcp.PdfProcessorClient
}

func NewPdfClient(endpoint string) *PdfClient {
	//创建gRPC客户端连接
	conn := zrpc.MustNewClient(zrpc.RpcClientConf{
		Endpoints: []string{endpoint},
		NonBlock:  true,
	})

	return &PdfClient{
		client: mcp.NewPdfProcessorClient(conn.Conn()),
	}
}

// ExtractText 简化接口：传入文件流和文件名，返回解析后的文本
func (c *PdfClient) ExtractText(file multipart.File, filename string) (string, error) {
	//创建grpc流
	stream, err := c.client.ExtractText(context.Background())
	if err != nil {
		logx.Errorf("gRPC连接失败：%v", err)
		return "", err
	}
	defer func() {
		if err := stream.CloseSend(); err != nil {
			logx.Errorf("关闭gRPC流失败%v", err)
		}
	}()

	//发送元数据
	if err := stream.Send(&mcp.PdfRequest{
		Data: &mcp.PdfRequest_Metadata{
			Metadata: &mcp.Metadata{
				Filename: filename,
				MimeType: "application/pdf",
			},
		},
	}); err != nil {
		logx.Errorf("读取文件失败%v", err)
		return "", err
	}

	//一次性发送整个文件（小文件直接发送）
	fileData, err := ioutil.ReadAll(file)
	if err != nil {
		logx.Errorf("读取文件失败：%v", err)
		return "", err
	}

	if err := stream.Send(&mcp.PdfRequest{
		Data: &mcp.PdfRequest_Chunk{
			Chunk: fileData,
		},
	}); err != nil {
		logx.Errorf("发送文件数据失败：%v", err)
		return "", err
	}

	//关闭发送并接收响应
	resp, err := stream.CloseAndRecv()
	if err != nil {
		logx.Errorf("接收响应失败：%v", err)
		return "", err
	}

	if resp.Error != "" {
		logx.Errorf("PDF解析错误：%s", resp.Error)
		return "", errors.New(resp.Error)
	}

	return resp.Content, nil
}
