package mongo

import (
	"context"
	"cv2/internal/infra/algorithm"
	"cv2/internal/infra/mongo/model"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const resumeContentCollection = "resume_content"

// SaveResumeContent 保存简历内容到 MongoDB
func (c *Client) SaveResumeContent(ctx context.Context, resumeID int64, data *algorithm.ResumeData) error {
	collection := c.Database().Collection(resumeContentCollection)

	// 构造 MongoDB 文档
	doc := &model.ResumeContent{
		MySQLID:    resumeID,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
		RawMD:      "", // TODO: 如果有 Markdown 原文可以存储
		Modules:    buildModules(data),
	}

	// 插入文档
	_, err := collection.InsertOne(ctx, doc)
	return err
}

// buildModules 构建模块数据
func buildModules(data *algorithm.ResumeData) []model.ModuleData {
	var modules []model.ModuleData

	// 基本信息模块
	if data.Name != "" || data.Phone != "" || data.Email != "" {
		basicInfo := map[string]interface{}{
			"name":        data.Name,
			"phone":       data.Phone,
			"email":       data.Email,
			"job_title":   data.JobTitle,
			"birthday":    data.Birthday,
			"ethnicity":   data.Ethnicity,
			"politics":    data.Politics,
			"location":    data.Location,
			"target_city": data.TargetCity,
			"min_salary":  data.MinSalary,
			"max_salary":  data.MaxSalary,
			"job_type":    data.JobType,
		}
		modules = append(modules, model.ModuleData{
			ModuleID: 1, // TODO: 从配置或数据库获取实际的模块ID
			Title:    "基本信息",
			Data:     []map[string]interface{}{basicInfo},
		})
	}

	// 教育经历模块
	if len(data.Education) > 0 {
		eduData := make([]map[string]interface{}, 0, len(data.Education))
		for _, edu := range data.Education {
			eduData = append(eduData, map[string]interface{}{
				"school_name": edu.SchoolName,
				"degree":      edu.Degree,
				"major":       edu.Major,
				"start_time":  edu.StartTime,
				"end_time":    edu.EndTime,
				"description": edu.Description,
			})
		}
		modules = append(modules, model.ModuleData{
			ModuleID: 2,
			Title:    "教育经历",
			Data:     eduData,
		})
	}

	// 在校经历模块
	if len(data.CampusExp) > 0 {
		campusData := make([]map[string]interface{}, 0, len(data.CampusExp))
		for _, campus := range data.CampusExp {
			campusData = append(campusData, map[string]interface{}{
				"title":       campus.Title,
				"role":        campus.Role,
				"start_time":  campus.StartTime,
				"end_time":    campus.EndTime,
				"description": campus.Description,
			})
		}
		modules = append(modules, model.ModuleData{
			ModuleID: 3,
			Title:    "在校经历",
			Data:     campusData,
		})
	}

	// 实习经历模块
	if len(data.InternExp) > 0 {
		internData := make([]map[string]interface{}, 0, len(data.InternExp))
		for _, intern := range data.InternExp {
			internData = append(internData, map[string]interface{}{
				"company":     intern.Company,
				"position":    intern.Position,
				"start_time":  intern.StartTime,
				"end_time":    intern.EndTime,
				"description": intern.Description,
			})
		}
		modules = append(modules, model.ModuleData{
			ModuleID: 4,
			Title:    "实习经历",
			Data:     internData,
		})
	}

	// 项目经历模块
	if len(data.ProjectExp) > 0 {
		projectData := make([]map[string]interface{}, 0, len(data.ProjectExp))
		for _, project := range data.ProjectExp {
			projectData = append(projectData, map[string]interface{}{
				"project_name": project.ProjectName,
				"role":         project.Role,
				"start_time":   project.StartTime,
				"end_time":     project.EndTime,
				"description":  project.Description,
			})
		}
		modules = append(modules, model.ModuleData{
			ModuleID: 5,
			Title:    "项目经历",
			Data:     projectData,
		})
	}

	// 技能证书模块
	if data.Skills != "" {
		modules = append(modules, model.ModuleData{
			ModuleID: 6,
			Title:    "技能证书",
			Data: []map[string]interface{}{
				{"content": data.Skills},
			},
		})
	}

	// 自我评价模块
	if data.SelfEval != "" {
		modules = append(modules, model.ModuleData{
			ModuleID: 7,
			Title:    "自我评价",
			Data: []map[string]interface{}{
				{"content": data.SelfEval},
			},
		})
	}

	return modules
}

// GetResumeContent 获取简历内容
func (c *Client) GetResumeContent(ctx context.Context, resumeID int64) (*model.ResumeContent, error) {
	collection := c.Database().Collection(resumeContentCollection)

	var result model.ResumeContent
	err := collection.FindOne(ctx, bson.M{"mysql_id": resumeID}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &result, nil
}
