// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package resume

import (
	"context"
	"net/http"

	"cv2/internal/infra/ent/module"
	"cv2/internal/infra/ent/resumescore"
	"cv2/internal/pkg/errx"
	"cv2/internal/svc"
	"cv2/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveModuleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 保存简历模块
func NewSaveModuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveModuleLogic {
	return &SaveModuleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SaveModuleLogic) SaveModule(req *types.SaveModuleReq) (resp *types.SaveModuleResp, err error) {
	// 1. 查询模块信息
	mod, err := l.svcCtx.Ent.Module.Query().
		Where(module.ID(req.ModuleID)).
		WithDimensions().
		First(l.ctx)
	if err != nil {
		return nil, errx.Warp(http.StatusNotFound, err, "模块不存在")
	}

	// 获取模块下的维度ID列表
	dimIDs := make([]int64, 0, len(mod.Edges.Dimensions))
	for _, dim := range mod.Edges.Dimensions {
		dimIDs = append(dimIDs, dim.ID)
	}

	// 2. 开启事务
	tx, err := l.svcCtx.Ent.Tx(l.ctx)
	if err != nil {
		return nil, errx.Warp(http.StatusInternalServerError, err, "开启事务失败")
	}
	defer tx.Rollback()

	// 3. 删除该模块的旧评分（模块得分 + 维度得分）
	_, err = tx.ResumeScore.Delete().
		Where(
			resumescore.ResumeID(req.ResumeID),
			resumescore.Or(
				resumescore.And(
					resumescore.TargetType(0),
					resumescore.TargetID(req.ModuleID),
				),
				resumescore.And(
					resumescore.TargetType(1),
					resumescore.TargetIDIn(dimIDs...),
				),
			),
		).Exec(l.ctx)
	if err != nil {
		return nil, errx.Warp(http.StatusInternalServerError, err, "删除旧评分失败")
	}

	// 4. 更新 MongoDB 中的模块数据
	err = l.svcCtx.Mongo.UpdateModule(l.ctx, req.ResumeID, req.ModuleID, mod.Title, req.Data)
	if err != nil {
		return nil, errx.Warp(http.StatusInternalServerError, err, "更新模块数据失败")
	}

	// 5. 重新计算该模块的评分
	calculator := NewScoreCalculator(l.ctx, l.svcCtx.Ent, l.svcCtx.Algorithm)
	moduleData := l.convertToModuleData(mod.Title, req.Data)
	err = calculator.ScoreSingleModule(tx, req.ResumeID, mod.Title, moduleData)
	if err != nil {
		l.Errorf("score module failed: %v", err)
		// 评分失败不影响保存，继续提交
	}

	// 6. 提交事务
	if err := tx.Commit(); err != nil {
		return nil, errx.Warp(http.StatusInternalServerError, err, "提交事务失败")
	}

	// 7. 查询新的评分结果
	scores, err := l.svcCtx.Ent.ResumeScore.Query().
		Where(
			resumescore.ResumeID(req.ResumeID),
			resumescore.Or(
				resumescore.And(
					resumescore.TargetType(0),
					resumescore.TargetID(req.ModuleID),
				),
				resumescore.And(
					resumescore.TargetType(1),
					resumescore.TargetIDIn(dimIDs...),
				),
			),
		).All(l.ctx)
	if err != nil {
		l.Errorf("query scores failed: %v", err)
	}

	// 构建得分映射
	var moduleScore float64
	dimScoreMap := make(map[int64]float64)
	for _, s := range scores {
		if s.TargetType == 0 {
			moduleScore = s.Score
		} else {
			dimScoreMap[s.TargetID] = s.Score
		}
	}

	// 构建维度得分列表
	var dimensions []types.DimensionInfo
	for _, dim := range mod.Edges.Dimensions {
		dimensions = append(dimensions, types.DimensionInfo{
			DimensionID: dim.ID,
			Title:       dim.Title,
			Score:       dimScoreMap[dim.ID],
		})
	}

	return &types.SaveModuleResp{
		ModuleID:   req.ModuleID,
		Title:      mod.Title,
		Score:      moduleScore,
		Data:       req.Data,
		Dimensions: dimensions,
	}, nil
}

// convertToModuleData 将请求数据转换为评分所需的格式
func (l *SaveModuleLogic) convertToModuleData(moduleTitle string, data []map[string]interface{}) interface{} {
	// 根据模块类型返回不同格式的数据
	switch moduleTitle {
	case "基本信息":
		if len(data) > 0 {
			return data[0]
		}
		return map[string]interface{}{}
	case "技能证书", "自我评价":
		if len(data) > 0 {
			if content, ok := data[0]["content"].(string); ok {
				return content
			}
		}
		return ""
	default:
		return data
	}
}
