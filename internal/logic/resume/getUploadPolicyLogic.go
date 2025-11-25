package resume

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"cv2/internal/pkg/snowflake"
	"cv2/internal/svc"
	"cv2/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUploadPolicyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取简历上传 POST Policy
func NewGetUploadPolicyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUploadPolicyLogic {
	return &GetUploadPolicyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUploadPolicyLogic) GetUploadPolicy(req *types.UploadPolicyReq) (resp *types.UploadPolicyResp, err error) {
	// 生成唯一的对象键
	fileExt := filepath.Ext(req.FileName)
	if fileExt == "" {
		fileExt = "." + req.FileType
	}

	// 使用雪花ID生成唯一文件名
	objectKey := fmt.Sprintf("resumes/%d%s", snowflake.NextID(), fileExt)

	// 设置过期时间为15分钟
	expiresIn := 15 * time.Minute

	// 设置最大文件大小，默认为 100MB
	maxFileSize := req.MaxFileSize
	if maxFileSize == 0 {
		maxFileSize = 100 * 1024 * 1024 // 100MB
	}

	// 获取预签名 POST Policy
	_, formData, err := l.svcCtx.MinIO.GetPresignedPostPolicy(l.ctx, objectKey, expiresIn, maxFileSize)
	if err != nil {
		logx.Errorf("Failed to get presigned post policy: %v", err)
		return nil, err
	}

	// 构造响应数据
	data := map[string]interface{}{
		"url":        formData["url"],
		"form_data":  formData,
		"object_key": objectKey,
		"expires_in": int64(expiresIn.Seconds()),
	}

	// 序列化为 JSON 字符串
	dataJSON, err := json.Marshal(data)
	if err != nil {
		logx.Errorf("Failed to marshal data: %v", err)
		return nil, err
	}

	return &types.UploadPolicyResp{
		Data: string(dataJSON),
	}, nil
}
