package resume

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"cv2/internal/pkg/snowflake"
	"cv2/internal/svc"
	"cv2/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUploadPresignURLLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取简历上传预签名 URL
func NewGetUploadPresignURLLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUploadPresignURLLogic {
	return &GetUploadPresignURLLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUploadPresignURLLogic) GetUploadPresignURL(req *types.UploadPresignReq) (resp *types.UploadPresignResp, err error) {
	// 生成唯一的对象键
	fileExt := filepath.Ext(req.FileName)
	if fileExt == "" {
		fileExt = "." + req.FileType
	}

	// 使用雪花ID生成唯一文件名
	objectKey := fmt.Sprintf("resumes/%d%s", snowflake.NextID(), fileExt)

	// 设置过期时间为15分钟
	expiresIn := 15 * time.Minute

	// 获取预签名上传 URL
	uploadURL, err := l.svcCtx.MinIO.GetPresignedUploadURL(l.ctx, objectKey, expiresIn)
	if err != nil {
		logx.Errorf("Failed to get presigned upload URL: %v", err)
		return nil, err
	}

	return &types.UploadPresignResp{
		UploadURL: uploadURL,
		ObjectKey: objectKey,
		ExpiresIn: int64(expiresIn.Seconds()),
	}, nil
}
