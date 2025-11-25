package algorithm

import (
	"fmt"
	"strings"
)

// QuestionAnswer 问题答案对
type QuestionAnswer struct {
	Question string   `json:"question"`
	Answers  []string `json:"answers"`
}

// BuildGenerateRequest 构造算法生成请求（简化版）
// 保留硬编码问题索引和 [MASK] 占位符逻辑
func BuildGenerateRequest(qas []QuestionAnswer) *GenerateRequest {
	// 构造 Info 字段（拼接所有问答）
	var infoBuilder strings.Builder
	for _, qa := range qas {
		infoBuilder.WriteString(qa.Question)
		infoBuilder.WriteString(strings.Join(qa.Answers, ","))
		infoBuilder.WriteByte('\n')
	}

	// 构造 Data 字段
	data := &ResumeData{
		Name:     "姓名",
		Phone:    "12345678901",
		Email:    "morgenAI",
		JobTitle: getJobTitle(qas),
	}

	// 教育经历（qas[0] - 专业）
	data.Education = buildEducation(qas)

	// 在校经历（qas[2] - 校园活动）
	data.CampusExp = buildCampusExp(qas)

	// 实习经历（基于意向岗位）
	data.InternExp = buildInternExp(qas)

	// 项目经历（初始化为空）
	data.ProjectExp = []ProjectExp{}

	// 技能和自评（使用 MASK 占位符）
	data.Skills = Mask
	data.SelfEval = Mask

	return &GenerateRequest{
		Data: data,
		Info: infoBuilder.String(),
	}
}

// getJobTitle 获取意向岗位（qas[1]）
func getJobTitle(qas []QuestionAnswer) string {
	if len(qas) > 1 && len(qas[1].Answers) > 0 {
		return qas[1].Answers[len(qas[1].Answers)-1]
	}
	return ""
}

// buildEducation 构造教育经历（qas[0] - 专业）
func buildEducation(qas []QuestionAnswer) []EducationExp {
	if len(qas) == 0 || len(qas[0].Answers) == 0 {
		return nil
	}

	var result []EducationExp
	for _, major := range qas[0].Answers {
		result = append(result, EducationExp{
			SchoolName:  "xx大学",
			Major:       major,
			StartTime:   "2021-09",
			EndTime:     "2025-06",
			Description: Mask,
		})
	}
	return result
}

// buildCampusExp 构造在校经历（qas[2] - 校园活动，最多4条）
func buildCampusExp(qas []QuestionAnswer) []CampusExp {
	if len(qas) < 3 || len(qas[2].Answers) == 0 {
		return nil
	}

	var result []CampusExp
	year := 2021
	maxCount := min(4, len(qas[2].Answers))

	for i := 0; i < maxCount; i++ {
		result = append(result, CampusExp{
			Title:       qas[2].Answers[i],
			Role:        Mask,
			StartTime:   fmt.Sprintf("%d.09", year),
			EndTime:     fmt.Sprintf("%d.06", year+1),
			Description: Mask,
		})
		year++
	}
	return result
}

// buildInternExp 构造实习经历（基于意向岗位）
func buildInternExp(qas []QuestionAnswer) []InternExp {
	jobTitle := getJobTitle(qas)
	if jobTitle == "" {
		return nil
	}

	return []InternExp{
		{
			Company:     "xxx公司",
			Position:    jobTitle + "实习生",
			StartTime:   "2021.09",
			EndTime:     "2022.06",
			Description: Mask,
		},
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
