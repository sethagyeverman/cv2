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
				updateRedisStatus(ctx, l.svcCtx, resumeTaskKeyPrefix, taskID, statusFailure, 0)
				l.Errorf("task failed: task_id=%s", taskID)
				return
			}

		case <-timeout:
			updateRedisStatus(ctx, l.svcCtx, resumeTaskKeyPrefix, taskID, statusTimeout, 0)
			l.Errorf("task timeout: task_id=%s", taskID)
			return
		}
	}
}

// handleSuccess 处理成功
func (l *GenerateResumeLogic) handleSuccess(
	ctx context.Context,
	taskID string,
	resumeID, userID, tenantID int64,
	data *algorithm.ResumeData,
) {
	// 1. 保存简历数据和评分
	err := saveAndProcessResume(ctx, l.svcCtx, resumeID, userID, tenantID, nil, data)
	if err != nil {
		l.Errorf("save resume data failed: %v", err)
		updateRedisStatus(ctx, l.svcCtx, resumeTaskKeyPrefix, taskID, statusFailure, 0)
		return
	}

	// 2. 更新 Redis 为 SUCCESS
	updateRedisStatus(ctx, l.svcCtx, resumeTaskKeyPrefix, taskID, statusSuccess, resumeID)
	l.Infof("resume generation completed: task_id=%s, resume_id=%d", taskID, resumeID)
}
