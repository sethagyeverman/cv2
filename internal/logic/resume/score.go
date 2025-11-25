package resume

import (
	"context"
	"fmt"
	"net/http"

	"cv2/internal/infra/algorithm"
	"cv2/internal/infra/ent"
	"cv2/internal/pkg/errx"

	"github.com/zeromicro/go-zero/core/logx"
)

// 简历模块定义
const (
	ModuleBasicInfo = "基本信息"
	ModuleEducation = "教育背景"
	ModuleCampus    = "在校经历"
	ModuleIntern    = "实习经历"
	ModuleProject   = "项目经历"
	ModuleSkills    = "技能证书"
	ModuleSelfEval  = "自我评价"
)

// 评分维度定义（简化版，实际应从配置或数据库读取）
var defaultScoringRules = map[string][]algorithm.RuleItem{
	"完整性": {
		{JudgmentDetail: "信息完整", JudgmentScore: 100},
		{JudgmentDetail: "信息较完整", JudgmentScore: 80},
		{JudgmentDetail: "信息不完整", JudgmentScore: 60},
	},
	"专业性": {
		{JudgmentDetail: "专业术语准确", JudgmentScore: 100},
		{JudgmentDetail: "专业术语较准确", JudgmentScore: 80},
		{JudgmentDetail: "专业术语不准确", JudgmentScore: 60},
	},
	"匹配度": {
		{JudgmentDetail: "高度匹配", JudgmentScore: 100},
		{JudgmentDetail: "基本匹配", JudgmentScore: 80},
		{JudgmentDetail: "匹配度低", JudgmentScore: 60},
	},
}

// ScoreCalculator 评分计算器
type ScoreCalculator struct {
	logx.Logger
	ctx       context.Context
	entClient *ent.Client
	algClient *algorithm.Client
}

// NewScoreCalculator 创建评分计算器
func NewScoreCalculator(ctx context.Context, entClient *ent.Client, algClient *algorithm.Client) *ScoreCalculator {
	return &ScoreCalculator{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		entClient: entClient,
		algClient: algClient,
	}
}

// CalculateResumeScore 计算整份简历的评分
func (c *ScoreCalculator) CalculateResumeScore(resumeID int64, data *algorithm.ResumeData) error {
	// 定义需要评分的模块
	modules := []struct {
		name string
		data interface{}
	}{
		{ModuleBasicInfo, c.extractBasicInfo(data)},
		{ModuleEducation, data.Education},
		{ModuleCampus, data.CampusExp},
		{ModuleIntern, data.InternExp},
		{ModuleProject, data.ProjectExp},
		{ModuleSkills, data.Skills},
		{ModuleSelfEval, data.SelfEval},
	}

	// 并发计算每个模块的评分
	for _, module := range modules {
		if err := c.scoreModule(resumeID, module.name, module.data); err != nil {
			c.Errorf("score module failed: module=%s, error=%v", module.name, err)
			// 单个模块评分失败不影响其他模块
			continue
		}
	}

	return nil
}

// scoreModule 对单个模块进行评分
func (c *ScoreCalculator) scoreModule(resumeID int64, moduleName string, moduleData interface{}) error {
	// 1. 准备评分请求
	section := map[string]interface{}{
		moduleName: moduleData,
	}

	req := &algorithm.ScoreRequest{
		Section: section,
		Rules:   defaultScoringRules,
	}

	// 2. 调用算法服务评分
	scores, err := c.algClient.ScoreResume(c.ctx, req)
	if err != nil {
		return fmt.Errorf("call algorithm service: %w", err)
	}

	// 3. 保存评分结果
	return c.saveModuleScores(resumeID, moduleName, scores)
}

// saveModuleScores 保存模块评分
func (c *ScoreCalculator) saveModuleScores(resumeID int64, moduleName string, scores []*algorithm.DimScore) error {
	if len(scores) == 0 {
		return nil
	}

	// 计算模块总分（加权平均）
	var totalScore float64
	var totalWeight float64 = float64(len(scores))

	for _, score := range scores {
		totalScore += score.Score
	}

	moduleScore := totalScore / totalWeight

	// 开启事务
	tx, err := c.entClient.Tx(c.ctx)
	if err != nil {
		return errx.Warp(http.StatusInternalServerError, err, "开启事务失败")
	}
	defer tx.Rollback()

	// 保存模块总分（target_type=0 表示模块）
	_, err = tx.ResumeScore.Create().
		SetResumeID(resumeID).
		SetTargetID(c.getModuleID(moduleName)).
		SetTargetType(0).
		SetScore(moduleScore).
		SetWeight(1.0).
		Save(c.ctx)
	if err != nil {
		return errx.Warp(http.StatusInternalServerError, err, "保存模块得分失败")
	}

	// 保存各维度得分（target_type=1 表示维度）
	for i, score := range scores {
		_, err := tx.ResumeScore.Create().
			SetResumeID(resumeID).
			SetTargetID(int64(i + 1)). // 维度ID，实际应该从配置获取
			SetTargetType(1).
			SetScore(score.Score).
			SetWeight(1.0 / totalWeight).
			Save(c.ctx)
		if err != nil {
			c.Errorf("save dimension score failed: dimension=%s, error=%v", score.Rule, err)
			continue
		}
	}

	return tx.Commit()
}

// extractBasicInfo 提取基本信息
func (c *ScoreCalculator) extractBasicInfo(data *algorithm.ResumeData) map[string]interface{} {
	return map[string]interface{}{
		"姓名":     data.Name,
		"电话":     data.Phone,
		"邮箱":     data.Email,
		"意向岗位":   data.JobTitle,
		"所在地":    data.Location,
		"意向城市":   data.TargetCity,
		"期望薪资上限": data.MaxSalary,
		"期望薪资下限": data.MinSalary,
	}
}

// getModuleID 获取模块ID（简化版，实际应该从配置或数据库获取）
func (c *ScoreCalculator) getModuleID(moduleName string) int64 {
	moduleIDs := map[string]int64{
		ModuleBasicInfo: 1,
		ModuleEducation: 2,
		ModuleCampus:    3,
		ModuleIntern:    4,
		ModuleProject:   5,
		ModuleSkills:    6,
		ModuleSelfEval:  7,
	}

	if id, ok := moduleIDs[moduleName]; ok {
		return id
	}
	return 0
}

// GetResumeScores 获取简历评分
func (c *ScoreCalculator) GetResumeScores(resumeID int64) (map[string]interface{}, error) {
	// 使用 ent 的查询方法需要导入 resumescore 包
	// 这里先简化实现，返回基本结构
	result := map[string]interface{}{
		"resume_id": resumeID,
		"modules":   []map[string]interface{}{},
	}

	// TODO: 实现完整的评分查询逻辑
	// scores, err := c.entClient.ResumeScore.Query().
	// 	Where(resumescore.ResumeIDEQ(resumeID)).
	// 	All(c.ctx)

	return result, nil
}

// ModuleScoreData 模块评分数据
type ModuleScoreData struct {
	ModuleID   int64                `json:"module_id"`
	ModuleName string               `json:"module_name"`
	Score      float64              `json:"score"`
	Dimensions []DimensionScoreData `json:"dimensions"`
}

// DimensionScoreData 维度评分数据
type DimensionScoreData struct {
	DimensionName string  `json:"dimension_name"`
	Score         float64 `json:"score"`
	Weight        float64 `json:"weight"`
}
