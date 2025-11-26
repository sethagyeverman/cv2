package resume

import (
	"context"
	"fmt"
	"net/http"

	"cv2/internal/infra/algorithm"
	"cv2/internal/infra/ent"
	"cv2/internal/infra/ent/dimension"
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

// getScoringRulesByModuleID 根据模块ID从数据库获取评分规则
func (c *ScoreCalculator) getScoringRulesByModuleID(moduleID int64) (map[string][]algorithm.RuleItem, error) {
	// 查询该模块下的所有维度
	dimensions, err := c.entClient.Dimension.Query().
		Where(dimension.ModuleIDEQ(moduleID)).
		All(c.ctx)

	if err != nil {
		return nil, fmt.Errorf("query dimensions failed: %w", err)
	}

	if len(dimensions) == 0 {
		c.Infof("no dimensions found for module_id=%d, using default rules", moduleID)
		return c.getDefaultScoringRules(), nil
	}

	// 构造评分规则
	rules := make(map[string][]algorithm.RuleItem)
	for _, dim := range dimensions {
		// 从 judgment 字段解析规则项
		if len(dim.Judgment) == 0 {
			c.Errorf("dimension %s has no judgment data", dim.Title)
			continue
		}

		ruleItems := make([]algorithm.RuleItem, 0, len(dim.Judgment))
		for _, j := range dim.Judgment {
			detail, _ := j["detail"].(string)
			score, _ := j["score"].(float64)

			ruleItems = append(ruleItems, algorithm.RuleItem{
				JudgmentDetail: detail,
				JudgmentScore:  score,
			})
		}

		if len(ruleItems) > 0 {
			rules[dim.Title] = ruleItems
		}
	}

	c.Infof("loaded %d scoring rules for module_id=%d", len(rules), moduleID)
	return rules, nil
}

// getDefaultScoringRules 获取默认评分规则（当数据库中没有配置时使用）
func (c *ScoreCalculator) getDefaultScoringRules() map[string][]algorithm.RuleItem {
	// 默认通用评分规则（当数据库中没有配置时使用）
	return map[string][]algorithm.RuleItem{
		"内容完整性": {
			{JudgmentDetail: "内容完整详细", JudgmentScore: 100},
			{JudgmentDetail: "内容基本完整", JudgmentScore: 80},
			{JudgmentDetail: "内容不够完整", JudgmentScore: 60},
		},
		"内容专业性": {
			{JudgmentDetail: "专业术语准确，表述规范", JudgmentScore: 100},
			{JudgmentDetail: "表述基本专业", JudgmentScore: 80},
			{JudgmentDetail: "表述不够专业", JudgmentScore: 60},
		},
	}
}

// ScoreCalculator 评分计算器
type ScoreCalculator struct {
	logx.Logger
	ctx       context.Context
	entClient *ent.Client
	algClient *algorithm.Client
	// 缓存模块ID到模块名称的映射
	moduleCache map[int64]string
}

// NewScoreCalculator 创建评分计算器
func NewScoreCalculator(ctx context.Context, entClient *ent.Client, algClient *algorithm.Client) *ScoreCalculator {
	return &ScoreCalculator{
		Logger:      logx.WithContext(ctx),
		ctx:         ctx,
		entClient:   entClient,
		algClient:   algClient,
		moduleCache: make(map[int64]string),
	}
}

// CalculateResumeScore 计算整份简历的评分（在事务中执行）
func (c *ScoreCalculator) CalculateResumeScore(tx *ent.Tx, resumeID int64, data *algorithm.ResumeData) error {
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

	// 计算每个模块的评分
	for _, module := range modules {
		if err := c.scoreModule(tx, resumeID, module.name, module.data); err != nil {
			c.Errorf("score module failed: module=%s, error=%v", module.name, err)
			// 单个模块评分失败不影响其他模块，继续执行
			continue
		}
	}

	return nil
}

// scoreModule 对单个模块进行评分
func (c *ScoreCalculator) scoreModule(tx *ent.Tx, resumeID int64, moduleName string, moduleData interface{}) error {
	// 1. 获取模块ID
	moduleID := c.getModuleID(moduleName)
	if moduleID == 0 {
		c.Errorf("unknown module name: %s, skip scoring", moduleName)
		return nil
	}

	// 2. 从数据库获取该模块的评分规则
	rules, err := c.getScoringRulesByModuleID(moduleID)
	if err != nil {
		return fmt.Errorf("get scoring rules failed: %w", err)
	}

	if len(rules) == 0 {
		c.Errorf("no scoring rules for module: %s (id=%d)", moduleName, moduleID)
		return nil
	}

	// 3. 准备评分请求
	section := map[string]interface{}{
		moduleName: moduleData,
	}

	req := &algorithm.ScoreRequest{
		Section: section,
		Rules:   rules,
	}

	// 4. 调用算法服务评分
	scores, err := c.algClient.ScoreResume(c.ctx, req)
	if err != nil {
		return fmt.Errorf("call algorithm service: %w", err)
	}

	// 4. 保存评分结果
	return c.saveModuleScores(tx, resumeID, moduleName, scores)
}

// saveModuleScores 保存模块评分（在事务中执行）
func (c *ScoreCalculator) saveModuleScores(tx *ent.Tx, resumeID int64, moduleName string, scores []*algorithm.DimScore) error {
	if len(scores) == 0 {
		c.Infof("no scores returned for module: %s", moduleName)
		return nil
	}

	// 计算模块总分（加权平均）
	var totalScore float64
	var totalWeight float64 = float64(len(scores))

	for _, score := range scores {
		totalScore += score.Score
	}

	moduleScore := totalScore / totalWeight

	// 保存模块总分（target_type=0 表示模块）
	moduleID := c.getModuleID(moduleName)
	_, err := tx.ResumeScore.Create().
		SetResumeID(resumeID).
		SetTargetID(moduleID).
		SetTargetType(0).
		SetScore(moduleScore).
		SetWeight(1.0).
		Save(c.ctx)
	if err != nil {
		return errx.Warpf(http.StatusInternalServerError, err, "保存模块得分失败: module=%s", moduleName)
	}

	c.Infof("saved module score: module=%s, score=%.2f", moduleName, moduleScore)

	// 保存各维度得分（target_type=1 表示维度）
	for i, score := range scores {
		// 使用模块ID和维度索引组合生成唯一的维度ID
		dimensionID := moduleID*100 + int64(i+1)

		_, err := tx.ResumeScore.Create().
			SetResumeID(resumeID).
			SetTargetID(dimensionID).
			SetTargetType(1).
			SetScore(score.Score).
			SetWeight(1.0 / totalWeight).
			Save(c.ctx)
		if err != nil {
			c.Errorf("save dimension score failed: dimension=%s, error=%v", score.Rule, err)
			continue
		}
		c.Infof("saved dimension score: module=%s, dimension=%s, score=%.2f", moduleName, score.Rule, score.Score)
	}

	return nil
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
