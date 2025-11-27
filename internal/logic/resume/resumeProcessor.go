package resume

import (
	"context"
	"mime/multipart"
	"net/http"
	"time"

	"cv2/internal/infra/algorithm"
	"cv2/internal/infra/minio"
	"cv2/internal/pkg/errx"
	"cv2/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	resumeTaskKeyPrefix = "resume_task:"
	statusPending       = "PENDING"
	statusProcessing    = "PROCESSING"
	statusSuccess       = "SUCCESS"
	statusFailure       = "FAILURE"
	statusTimeout       = "TIMEOUT"
)

// saveAndProcessResume 保存简历数据并异步处理文件
func saveAndProcessResume(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	resumeID, userID, tenantID int64,
	fileHeader *multipart.FileHeader,
	data *algorithm.ResumeData,
) error {
	// 1. 确定文件名
	fileName := data.Name + "-简历"
	if fileHeader != nil {
		fileName = fileHeader.Filename
	}

	// 2. 事务保存数据
	if err := saveResumeTransaction(ctx, svcCtx, resumeID, userID, tenantID, fileName, data); err != nil {
		return err
	}

	// 3. 异步处理文件
	go asyncProcessFile(context.Background(), svcCtx, resumeID, fileHeader, data)

	return nil
}

// saveResumeTransaction 在事务中保存简历数据
func saveResumeTransaction(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	resumeID, userID, tenantID int64,
	fileName string,
	data *algorithm.ResumeData,
) error {
	// 开启事务
	tx, err := svcCtx.Ent.Tx(ctx)
	if err != nil {
		return errx.Warp(http.StatusInternalServerError, err, "开启事务失败")
	}
	defer tx.Rollback()

	// 1. 创建简历主表记录
	_, err = tx.Resume.Create().
		SetID(resumeID).
		SetUserID(userID).
		SetTenantID(tenantID).
		SetFileName(fileName).
		SetCoverImage(svcCtx.Config.DefaultConfig.DefaultCoverImage).
		SetFilePath("uploading..."). // 异步操作完成后更新
		SetStatus(3).                // completed
		Save(ctx)
	if err != nil {
		return errx.Warp(http.StatusInternalServerError, err, "创建简历记录失败")
	}

	// 2. 保存到 MongoDB（简历详细内容）
	err = svcCtx.Mongo.SaveResumeContent(ctx, resumeID, data)
	if err != nil {
		return errx.Warp(http.StatusInternalServerError, err, "保存简历内容失败")
	}

	// 3. 计算并保存评分
	calculator := NewScoreCalculator(ctx, svcCtx.Ent, svcCtx.Algorithm)
	if err := calculator.CalculateResumeScore(tx, resumeID, data); err != nil {
		logx.WithContext(ctx).Errorf("calculate resume score failed: %v", err)
		// 评分失败不影响简历保存
	}

	// 4. 提交事务
	return tx.Commit()
}

// asyncProcessFile 异步处理文件（上传或生成）并更新数据库
func asyncProcessFile(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	resumeID int64,
	fileHeader *multipart.FileHeader,
	data *algorithm.ResumeData,
) {
	var (
		objectKey string
		err       error
		logger    = logx.WithContext(ctx)
	)

	if fileHeader != nil {
		// 模式1：上传已有文件
		objectKey = minio.GenerateObjectKey(fileHeader.Filename)
		if err = svcCtx.MinIO.Upload(ctx, objectKey, fileHeader); err != nil {
			logger.Errorf("async upload file to minio failed: resume_id=%d, err=%v", resumeID, err)
			return
		}
	} else {
		// 模式2：生成 PDF 文件 (GenerateResumeLogic)
		generator := NewPDFGenerator(svcCtx.Algorithm, svcCtx.MinIO)
		objectKey, err = generator.GenerateAndUpload(ctx, resumeID, data)
		if err != nil {
			logger.Errorf("generate and upload pdf failed: resume_id=%d, err=%v", resumeID, err)
			return
		}
	}

	// 更新简历表的 file_path
	fileURL := svcCtx.MinIO.GetPublicURL(objectKey)
	_, err = svcCtx.Ent.Resume.UpdateOneID(resumeID).
		SetFilePath(fileURL).
		Save(ctx)
	if err != nil {
		logger.Errorf("update resume file_path failed: resume_id=%d, err=%v", resumeID, err)
		return
	}

	logger.Infof("file processed asynchronously: resume_id=%d, url=%s", resumeID, fileURL)
}

// updateRedisStatus 更新 Redis 状态
func updateRedisStatus(ctx context.Context, svcCtx *svc.ServiceContext, keyPrefix, taskID, status string, resumeID int64) {
	key := keyPrefix + taskID
	svcCtx.Redis.HSet(ctx, key, "status", status)
	if resumeID > 0 {
		svcCtx.Redis.HSet(ctx, key, "resume_id", resumeID)
	}
	svcCtx.Redis.Expire(ctx, key, 12*time.Hour)
}
