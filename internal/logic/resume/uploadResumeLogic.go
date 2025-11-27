// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package resume

import (
	"context"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"cv2/internal/pkg/errx"
	"cv2/internal/pkg/snowflake"
	"cv2/internal/svc"
	"cv2/internal/types"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	maxFileSize = 20 * 1024 * 1024 // 20MB
)

type UploadResumeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 上传简历文件解析
func NewUploadResumeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadResumeLogic {
	return &UploadResumeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadResumeLogic) UploadResume(fileHeader *multipart.FileHeader) (resp *types.UploadResumeResp, err error) {
	// TODO: 从 context 获取 userID 和 tenantID
	userID := int64(1)
	tenantID := int64(1)

	// 1. 校验文件大小
	if fileHeader.Size > maxFileSize {
		return nil, errx.New(http.StatusBadRequest, "文件过大，请压缩后上传（最大20MB）")
	}

	// 2. 校验文件扩展名
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	allowedExts := map[string]bool{".pdf": true, ".doc": true, ".docx": true}
	if !allowedExts[ext] {
		return nil, errx.New(http.StatusBadRequest, "请上传 pdf/doc/docx 格式的简历")
	}

	// 3. 预生成简历ID
	resumeID := snowflake.NextID()

	// 4. 生成 taskID
	taskID := uuid.New().String()

	// 5. 写入 Redis（状态: PENDING）
	key := resumeTaskKeyPrefix + taskID
	err = l.svcCtx.Redis.HSet(l.ctx, key, map[string]any{
		"status":    statusPending,
		"resume_id": resumeID,
		"user_id":   userID,
		"tenant_id": tenantID,
	}).Err()
	if err != nil {
		l.Errorf("redis hset failed: %v", err)
		return nil, errx.Warp(http.StatusInternalServerError, err, "缓存任务失败")
	}

	// 设置过期时间
	l.svcCtx.Redis.Expire(l.ctx, key, 12*time.Hour)

	// 6. 启动后台处理
	go l.processUpload(context.Background(), taskID, resumeID, userID, tenantID, fileHeader)

	l.Infof("upload task created: task_id=%s, resume_id=%d", taskID, resumeID)

	return &types.UploadResumeResp{
		TaskID: taskID,
	}, nil
}

// processUpload 后台处理上传文件
func (l *UploadResumeLogic) processUpload(
	ctx context.Context,
	taskID string,
	resumeID, userID, tenantID int64,
	fileHeader *multipart.FileHeader,
) {
	// 1. 文件转 Markdown
	l.Infof("start file to markdown: resume_id=%d, filename=%s", resumeID, fileHeader.Filename)
	md, err := l.svcCtx.Algorithm.File2Markdown(ctx, fileHeader)
	if err != nil {
		l.Errorf("file to markdown failed: %v", err)
		updateRedisStatus(ctx, l.svcCtx, resumeTaskKeyPrefix, taskID, statusFailure, 0)
		return
	}
	l.Infof("file to markdown success: resume_id=%d, md_length=%d", resumeID, len(md))

	// 2. Markdown 转结构化数据
	l.Infof("start markdown to struct: resume_id=%d", resumeID)
	data, err := l.svcCtx.Algorithm.Markdown2Struct(ctx, md)
	if err != nil {
		l.Errorf("markdown to struct failed: %v", err)
		updateRedisStatus(ctx, l.svcCtx, resumeTaskKeyPrefix, taskID, statusFailure, 0)
		return
	}
	l.Infof("markdown to struct success: resume_id=%d", resumeID)

	// 3. 保存上传的简历
	if err := saveAndProcessResume(ctx, l.svcCtx, resumeID, userID, tenantID, fileHeader, data); err != nil {
		l.Errorf("save upload resume failed: %v", err)
		updateRedisStatus(ctx, l.svcCtx, resumeTaskKeyPrefix, taskID, statusFailure, 0)
		return
	}

	// 4. 更新 Redis 为 SUCCESS
	updateRedisStatus(ctx, l.svcCtx, resumeTaskKeyPrefix, taskID, statusSuccess, resumeID)
	l.Infof("upload resume completed: task_id=%s, resume_id=%d", taskID, resumeID)
}
