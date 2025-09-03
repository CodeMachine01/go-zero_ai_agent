package utils

import (
	"bytes"
	"github.com/unidoc/unipdf/v3/extractor"
	"github.com/unidoc/unipdf/v3/model"
	"io"
	"strings"
)

// ExtractPDFText 从io.Reader提取文本
func ExtractPDFText(file io.Reader) (string, error) {
	//创建内存缓冲去避免重复读取
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		return "", err
	}

	pdfReader, err := model.NewPdfReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		return "", err
	}
	var textBuilder strings.Builder
	numsPages, err := pdfReader.GetNumPages()
	if err != nil {
		return "", err
	}

	for i := 1; i <= numsPages; i++ {
		page, err := pdfReader.GetPage(i)
		if err != nil {
			return "", err
		}

		ex, err := extractor.New(page)
		if err != nil {
			return "", err
		}

		pageText, err := ex.ExtractText()
		if err != nil {
			return "", err
		}

		textBuilder.WriteString(strings.TrimSpace(pageText))
		textBuilder.WriteString("\n\n")
	}
	return textBuilder.String(), nil
}
