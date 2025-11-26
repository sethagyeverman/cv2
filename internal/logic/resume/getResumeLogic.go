// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package resume

import (
	"context"
	"net/http"

	"cv2/internal/infra/ent/resumescore"
	"cv2/internal/pkg/errx"
	"cv2/internal/svc"
	"cv2/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetResumeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取简历详情
func NewGetResumeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetResumeLogic {
	return &GetResumeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetResumeLogic) GetResume(req *types.GetResumeReq) (resp *types.GetResumeResp, err error) {
	// 1. 查询简历基本信息
	resume, err := l.svcCtx.Ent.Resume.Get(l.ctx, req.ResumeID)
	if err != nil {
		return nil, errx.Warp(http.StatusNotFound, err, "简历不存在")
	}

	// 2. 从 MongoDB 获取简历内容
	content, err := l.svcCtx.Mongo.GetResumeContent(l.ctx, req.ResumeID)
	if err != nil {
		return nil, errx.Warp(http.StatusInternalServerError, err, "获取简历内容失败")
	}

	// 3. 查询所有模块（含维度）
	modules, err := l.svcCtx.Ent.Module.Query().
		WithDimensions().
		All(l.ctx)
	if err != nil {
		return nil, errx.Warp(http.StatusInternalServerError, err, "获取模块信息失败")
	}

	// 4. 查询简历得分
	scores, err := l.svcCtx.Ent.ResumeScore.Query().
		Where(resumescore.ResumeID(req.ResumeID)).
		All(l.ctx)
	if err != nil {
		return nil, errx.Warp(http.StatusInternalServerError, err, "获取简历得分失败")
	}

	// 构建得分映射
	moduleScoreMap := make(map[int64]float64)
	dimScoreMap := make(map[int64]float64)
	for _, s := range scores {
		if s.TargetType == 0 { // 模块得分
			moduleScoreMap[s.TargetID] = s.Score
		} else { // 维度得分
			dimScoreMap[s.TargetID] = s.Score
		}
	}

	// 构建模块数据映射（从 MongoDB）
	moduleDataMap := make(map[int64][]map[string]interface{})
	if content != nil {
		for _, mod := range content.Modules {
			moduleDataMap[mod.ModuleID] = mod.Data
		}
	}

	// 5. 构建统一的模块列表（数据 + 得分）
	var moduleInfos []types.ModuleInfo
	var totalScore float64
	for _, m := range modules {
		mi := types.ModuleInfo{
			ModuleID:   m.ID,
			Title:      m.Title,
			Score:      moduleScoreMap[m.ID],
			Data:       moduleDataMap[m.ID],
			Dimensions: []types.DimensionInfo{},
		}
		totalScore += moduleScoreMap[m.ID]

		for _, dim := range m.Edges.Dimensions {
			mi.Dimensions = append(mi.Dimensions, types.DimensionInfo{
				DimensionID: dim.ID,
				Title:       dim.Title,
				Score:       dimScoreMap[dim.ID],
			})
		}
		moduleInfos = append(moduleInfos, mi)
	}

	return &types.GetResumeResp{
		ResumeID:   resume.ID,
		FileName:   resume.FileName,
		Status:     resume.Status,
		TotalScore: totalScore,
		Modules:    moduleInfos,
		CreatedAt:  resume.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:  resume.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}
