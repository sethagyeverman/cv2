package resume

import (
	"context"
	"fmt"

	"cv2/internal/infra/algorithm"
	"cv2/internal/infra/minio"

	"github.com/zeromicro/go-zero/core/logx"
)

// PDFGenerator PDF 生成器
type PDFGenerator struct {
	logx.Logger
	algorithm *algorithm.Client
	minio     *minio.Client
}

// NewPDFGenerator 创建 PDF 生成器
func NewPDFGenerator(algClient *algorithm.Client, minioClient *minio.Client) *PDFGenerator {
	return &PDFGenerator{
		Logger:    logx.WithContext(context.Background()),
		algorithm: algClient,
		minio:     minioClient,
	}
}

// GenerateAndUpload 根据结构化数据生成 PDF 并上传
// 返回 MinIO 中的 object key
func (g *PDFGenerator) GenerateAndUpload(ctx context.Context, resumeID int64, data *algorithm.ResumeData) (string, error) {
	// 1. 结构化数据转 Markdown（调用算法接口）
	md, err := g.algorithm.Struct2Markdown(ctx, data, data.Name+"-简历")
	if err != nil {
		return "", fmt.Errorf("struct to markdown: %w", err)
	}
	g.Infof("generated markdown for resume %d, length=%d", resumeID, len(md))

	// 2. Markdown 转 PDF
	filename := fmt.Sprintf("%s-简历.pdf", data.Name)
	pdfData, err := g.algorithm.Markdown2PDF(ctx, md, filename)
	if err != nil {
		return "", fmt.Errorf("markdown to pdf: %w", err)
	}
	g.Infof("generated pdf for resume %d, size=%d bytes", resumeID, len(pdfData))

	// 3. 上传到 MinIO
	objectKey := minio.GenerateObjectKey(filename)
	if err := g.minio.UploadBytes(ctx, objectKey, pdfData, "application/pdf"); err != nil {
		return "", fmt.Errorf("upload pdf: %w", err)
	}
	g.Infof("uploaded pdf for resume %d, key=%s", resumeID, objectKey)

	return objectKey, nil
}
