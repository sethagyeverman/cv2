package resume

import (
	"context"
	"cv2/internal/infra/algorithm"
	"cv2/internal/pkg/errx"
	"cv2/internal/pkg/snowflake"
	"cv2/internal/svc"
	"cv2/internal/types"
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	genTaskKeyPrefix = "gen_task:"
	statusPending    = "PENDING"
	statusProcessing = "PROCESSING"
	statusSuccess    = "SUCCESS"
	statusFailure    = "FAILURE"
	statusTimeout    = "TIMEOUT"
)

type GenerateResumeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 生成简历
func NewGenerateResumeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateResumeLogic {
	return &GenerateResumeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GenerateResumeLogic) GenerateResume(req *types.GenerateResumeReq) (resp *types.GenerateResumeResp, err error) {
	// TODO: 从 context 获取 userID 和 tenantID
	userID := int64(1)   // 临时硬编码
	tenantID := int64(1) // 临时硬编码

	// 1. 转换为 algorithm.QuestionAnswer
	qas := make([]algorithm.QuestionAnswer, len(req.Questions))
	for i, q := range req.Questions {
		qas[i] = algorithm.QuestionAnswer{
			Question: q.Question,
			Answers:  q.Answers,
		}
	}

	// 2. 构造算法请求
	algReq := algorithm.BuildGenerateRequest(qas)

	// 3. 提交任务到算法服务
	algResp, err := l.svcCtx.Algorithm.SubmitGenerateTask(l.ctx, algReq)
	if err != nil {
		l.Errorf("submit generate task failed: %v", err)
		return nil, errx.Warp(http.StatusInternalServerError, err, "提交任务失败")
	}
	taskID := algResp.TaskID

	// 4. 预生成简历ID
	resumeID := snowflake.NextID()

	// 5. 写入 Redis（状态: PENDING）
	key := genTaskKeyPrefix + taskID
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

	// 6. 启动后台监控
	go l.monitorTask(context.Background(), taskID, resumeID, userID, tenantID)

	l.Infof("resume generation task created: task_id=%s, resume_id=%d", taskID, resumeID)

	return &types.GenerateResumeResp{
		TaskID: taskID,
	}, nil
}

// monitorTask 监控任务状态
func (l *GenerateResumeLogic) monitorTask(
	ctx context.Context,
	taskID string,
	resumeID, userID, tenantID int64,
) {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	timeout := time.After(30 * time.Minute)

	for {
		select {
		case <-ticker.C:
			status, err := l.svcCtx.Algorithm.GetTaskStatus(ctx, taskID)
			if err != nil {
				l.Errorf("get task status failed: task_id=%s, err=%v", taskID, err)
				continue
			}

			l.Infof("task status: task_id=%s, status=%s", taskID, status.Status)

			switch status.Status {
			case "SUCCESS":
				l.handleSuccess(ctx, taskID, resumeID, userID, tenantID, status.Result)
				return
			case "FAILURE":
				l.updateRedisStatus(ctx, taskID, statusFailure, 0)
				l.Errorf("task failed: task_id=%s", taskID)
				return
			}

		case <-timeout:
			l.updateRedisStatus(ctx, taskID, statusTimeout, 0)
			l.Errorf("task timeout: task_id=%s", taskID)
			return
		}
	}
}

// handleSuccess 处理成功（事务 + 评分后更新状态）
func (l *GenerateResumeLogic) handleSuccess(
	ctx context.Context,
	taskID string,
	resumeID, userID, tenantID int64,
	data *algorithm.ResumeData,
) {
	// 1. 开启事务，保存简历数据和评分
	err := l.saveResumeDataWithScore(ctx, resumeID, userID, tenantID, data)
	if err != nil {
		l.Errorf("save resume data with score failed: %v", err)
		l.updateRedisStatus(ctx, taskID, statusFailure, 0)
		return
	}

	// 2. 更新 Redis 为 SUCCESS
	l.updateRedisStatus(ctx, taskID, statusSuccess, resumeID)
	l.Infof("resume generation completed: task_id=%s, resume_id=%d", taskID, resumeID)
}

// saveResumeDataWithScore 保存简历数据和评分（在同一个事务中）
func (l *GenerateResumeLogic) saveResumeDataWithScore(
	ctx context.Context,
	resumeID, userID, tenantID int64,
	data *algorithm.ResumeData,
) error {
	// 开启事务
	tx, err := l.svcCtx.Ent.Tx(ctx)
	if err != nil {
		return errx.Warp(http.StatusInternalServerError, err, "开启事务失败")
	}
	defer tx.Rollback()

	// 1. 创建简历主表记录
	_, err = tx.Resume.Create().
		SetID(resumeID).
		SetUserID(userID).
		SetTenantID(tenantID).
		SetFileName(data.Name + "-简历").
		SetFilePath("fakefakefakefkae"). // TODO: 生成文件后更新
		SetStatus(3).                    // completed
		Save(ctx)
	if err != nil {
		return errx.Warp(http.StatusInternalServerError, err, "创建简历记录失败")
	}

	// 2. 保存到 MongoDB（简历详细内容）
	err = l.svcCtx.Mongo.SaveResumeContent(ctx, resumeID, data)
	if err != nil {
		return errx.Warp(http.StatusInternalServerError, err, "保存简历内容失败")
	}

	// 3. 计算并保存评分
	calculator := NewScoreCalculator(ctx, l.svcCtx.Ent, l.svcCtx.Algorithm)
	if err := calculator.CalculateResumeScore(tx, resumeID, data); err != nil {
		l.Errorf("calculate resume score failed: %v", err)
		// 评分失败不影响简历保存，继续提交事务
		// 可以选择：1. 继续提交（当前方案） 2. 回滚事务
	}

	// 4. TODO: 保存到 MinIO（生成PDF文件）

	// 提交事务
	return tx.Commit()
}

// updateRedisStatus 更新 Redis 状态
func (l *GenerateResumeLogic) updateRedisStatus(ctx context.Context, taskID, status string, resumeID int64) {
	key := genTaskKeyPrefix + taskID
	l.svcCtx.Redis.HSet(ctx, key, "status", status)
	if resumeID > 0 {
		l.svcCtx.Redis.HSet(ctx, key, "resume_id", resumeID)
	}
	l.svcCtx.Redis.Expire(ctx, key, 12*time.Hour)
}
