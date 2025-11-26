// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package resume

import (
	"context"
	"cv2/internal/infra/algorithm"
	"cv2/internal/infra/ent/dimension"
	"cv2/internal/infra/ent/module"
	"cv2/internal/svc"
	"cv2/internal/types"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
)

type AIWriteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// AI 帮写（SSE 流式返回）
func NewAIWriteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AIWriteLogic {
	return &AIWriteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// AIWrite 执行 AI 帮写，将结果写入 client channel
func (l *AIWriteLogic) AIWrite(req *types.AIWriteReq, client chan<- string) error {
	l.Infof("AI write started: resume_id=%d, module_id=%d", req.ResumeID, req.ModuleID)

	// 构建带 [MASK] 的简历数据
	resumeData, err := l.buildMaskedResume(req.ResumeID, req.ModuleID)
	if err != nil {
		l.Errorf("build masked resume failed: %v", err)
		return err
	}

	// 构建 requirement
	requirement := l.buildRequirement(req.ModuleID)

	// 调用算法服务
	algReq := &algorithm.AIWriteRequest{
		Resume:      resumeData,
		Requirement: requirement,
		Info:        req.Info,
	}

	dataCh, errCh := l.svcCtx.Algorithm.AIWrite(l.ctx, algReq)

	for {
		select {
		case <-l.ctx.Done():
			return l.ctx.Err()
		case data, ok := <-dataCh:
			if !ok {
				return nil
			}
			client <- data
		case err, ok := <-errCh:
			if ok && err != nil {
				return err
			}
		}
	}
}

// buildMaskedResume 构建带 [MASK] 标记的简历数据
func (l *AIWriteLogic) buildMaskedResume(resumeID, moduleID int64) (*algorithm.ResumeData, error) {
	// 从 MongoDB 获取简历内容
	content, err := l.svcCtx.Mongo.GetResumeContent(l.ctx, resumeID)
	if err != nil {
		return nil, fmt.Errorf("get resume content: %w", err)
	}
	if content == nil {
		return nil, fmt.Errorf("resume not found: %d", resumeID)
	}

	// 转换为算法请求格式
	data := &algorithm.ResumeData{}

	for _, module := range content.Modules {
		isMask := module.ModuleID == moduleID
		switch module.Title {
		case "基本信息":
			l.fillBasicInfo(data, module.Data)
		case "教育背景":
			l.fillEducation(data, module.Data, isMask)
		case "在校经历":
			l.fillCampus(data, module.Data, isMask)
		case "实习经历":
			l.fillInternship(data, module.Data, isMask)
		case "项目经历":
			l.fillProject(data, module.Data, isMask)
		case "技能证书":
			l.fillSkill(data, module.Data, isMask)
		case "自我评价":
			l.fillSelfEval(data, module.Data, isMask)
		}
	}

	return data, nil
}

// buildRequirement 从模块关联的维度构建评分要求
func (l *AIWriteLogic) buildRequirement(moduleID int64) string {
	dimensions, err := l.svcCtx.Ent.Dimension.Query().
		Where(dimension.HasModuleWith(module.ID(moduleID))).
		All(l.ctx)

	if err != nil {
		return "时间、职务、活动名称、经历完整清晰"
	}

	var parts []string
	for _, dim := range dimensions {
		if dim.Description != "" {
			parts = append(parts, dim.Description)
		}
	}
	if len(parts) == 0 {
		return "时间、职务、活动名称、经历完整清晰"
	}
	return strings.Join(parts, "；")
}

// fillBasicInfo 填充基本信息
func (l *AIWriteLogic) fillBasicInfo(data *algorithm.ResumeData, items []map[string]interface{}) {
	if len(items) == 0 {
		return
	}
	item := items[0]
	data.Name = getString(item, "name")
	data.Phone = getString(item, "phone")
	data.Email = getString(item, "email")
	data.JobTitle = getString(item, "job_title")
	data.Birthday = getString(item, "birthday")
	data.Ethnicity = getString(item, "ethnicity")
	data.Politics = getString(item, "politics")
	data.Location = getString(item, "location")
	data.TargetCity = getString(item, "target_city")
	data.MinSalary = getString(item, "min_salary")
	data.MaxSalary = getString(item, "max_salary")
	data.JobType = getString(item, "job_type")
}

// fillEducation 填充教育经历
func (l *AIWriteLogic) fillEducation(data *algorithm.ResumeData, items []map[string]interface{}, isMask bool) {
	for _, item := range items {
		edu := algorithm.EducationExp{
			SchoolName:  getString(item, "school_name"),
			Degree:      getString(item, "degree"),
			Major:       getString(item, "major"),
			StartTime:   getString(item, "start_time"),
			EndTime:     getString(item, "end_time"),
			Description: getString(item, "description"),
		}
		if isMask {
			edu.Description = algorithm.Mask
		}
		data.Education = append(data.Education, edu)
	}
}

// fillCampus 填充在校经历
func (l *AIWriteLogic) fillCampus(data *algorithm.ResumeData, items []map[string]interface{}, isMask bool) {
	for _, item := range items {
		campus := algorithm.CampusExp{
			Title:       getString(item, "title"),
			Role:        getString(item, "role"),
			StartTime:   getString(item, "start_time"),
			EndTime:     getString(item, "end_time"),
			Description: getString(item, "description"),
		}
		if isMask {
			campus.Description = algorithm.Mask
		}
		data.CampusExp = append(data.CampusExp, campus)
	}
}

// fillInternship 填充实习经历
func (l *AIWriteLogic) fillInternship(data *algorithm.ResumeData, items []map[string]interface{}, isMask bool) {
	for _, item := range items {
		intern := algorithm.InternExp{
			Company:     getString(item, "company"),
			Position:    getString(item, "position"),
			StartTime:   getString(item, "start_time"),
			EndTime:     getString(item, "end_time"),
			Description: getString(item, "description"),
		}
		if isMask {
			intern.Description = algorithm.Mask
		}
		data.InternExp = append(data.InternExp, intern)
	}
}

// fillProject 填充项目经历
func (l *AIWriteLogic) fillProject(data *algorithm.ResumeData, items []map[string]interface{}, isMask bool) {
	for _, item := range items {
		project := algorithm.ProjectExp{
			ProjectName: getString(item, "project_name"),
			Role:        getString(item, "role"),
			StartTime:   getString(item, "start_time"),
			EndTime:     getString(item, "end_time"),
			Description: getString(item, "description"),
		}
		if isMask {
			project.Description = algorithm.Mask
		}
		data.ProjectExp = append(data.ProjectExp, project)
	}
}

// fillSkill 填充技能证书
func (l *AIWriteLogic) fillSkill(data *algorithm.ResumeData, items []map[string]interface{}, isMask bool) {
	if len(items) == 0 {
		return
	}
	data.Skills = getString(items[0], "content")
	if isMask {
		data.Skills = algorithm.Mask
	}
}

// fillSelfEval 填充自我评价
func (l *AIWriteLogic) fillSelfEval(data *algorithm.ResumeData, items []map[string]interface{}, isMask bool) {
	if len(items) == 0 {
		return
	}
	data.SelfEval = getString(items[0], "content")
	if isMask {
		data.SelfEval = algorithm.Mask
	}
}

// getString 从 map 中安全获取字符串
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
