package svc

import (
	"GoAgent/mcp/internal/config"
	"fmt"
	"github.com/unidoc/unipdf/v3/common/license"
)

type ServiceContext struct {
	Config config.Config
}

func NewServiceContext(c config.Config) *ServiceContext {
	//设置UniPDF key
	err := license.SetMeteredKey(c.UniPDFLicense)
	if err != nil {
		fmt.Printf("设置UniPDF许可证失败：: %v\n", err)
		//如果没有授权，UniPDF会添加水印
	}
	return &ServiceContext{
		Config: c,
	}
}
