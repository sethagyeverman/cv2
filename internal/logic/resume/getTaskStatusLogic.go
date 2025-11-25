// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package resume

import (
	"context"
	"cv2/internal/pkg/errx"
	"cv2/internal/svc"
	"cv2/internal/types"
	"net/http"
	"strconv"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetTaskStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 查询生成任务状态
func NewGetTaskStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTaskStatusLogic {
	return &GetTaskStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTaskStatusLogic) GetTaskStatus(req *types.TaskStatusReq) (resp *types.TaskStatusResp, err error) {
	taskID := req.TaskID
	key := genTaskKeyPrefix + taskID

	// 从 Redis 获取任务状态
	result, err := l.svcCtx.Redis.HGetAll(l.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, errx.New(http.StatusNotFound, "任务不存在")
		}
		l.Errorf("redis hgetall failed: %v", err)
		return nil, errx.Warp(http.StatusInternalServerError, err, "查询任务状态失败")
	}

	if len(result) == 0 {
		return nil, errx.New(http.StatusNotFound, "任务不存在")
	}

	// 构造响应
	resp = &types.TaskStatusResp{
		Status: result["status"],
	}

	// 如果有 resume_id，转换并返回
	if resumeIDStr, ok := result["resume_id"]; ok && resumeIDStr != "" {
		resumeID, err := strconv.ParseInt(resumeIDStr, 10, 64)
		if err == nil {
			resp.ResumeID = resumeID
		}
	}

	l.Infof("task status queried: task_id=%s, status=%s", taskID, resp.Status)

	return resp, nil
}
