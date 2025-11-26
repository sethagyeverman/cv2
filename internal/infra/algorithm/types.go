package algorithm

// Mask 占位符常量
const Mask = "[MASK]"

// GenerateRequest 简历生成请求
type GenerateRequest struct {
	Data *ResumeData `json:"data"`
	Info string      `json:"info"`
}

// GenerateResponse 简历生成响应
type GenerateResponse struct {
	TaskID string `json:"task_id"`
}

// TaskStatus 任务状态
type TaskStatus struct {
	Status string      `json:"task_status"`
	Result *ResumeData `json:"task_result,omitempty"`
}

// ResumeData 简历数据
type ResumeData struct {
	Name       string         `json:"姓名"`
	Phone      string         `json:"电话"`
	Email      string         `json:"邮箱"`
	JobTitle   string         `json:"意向岗位"`
	Birthday   string         `json:"出生日期,omitempty"`
	Ethnicity  string         `json:"民族,omitempty"`
	Politics   string         `json:"政治面貌,omitempty"`
	Location   string         `json:"所在地,omitempty"`
	TargetCity string         `json:"意向城市,omitempty"`
	MaxSalary  string         `json:"期望薪资上限,omitempty"`
	MinSalary  string         `json:"期望薪资下限,omitempty"`
	JobType    string         `json:"求职类型,omitempty"`
	Education  []EducationExp `json:"教育经历"`
	CampusExp  []CampusExp    `json:"在校经历"`
	InternExp  []InternExp    `json:"实习经历"`
	ProjectExp []ProjectExp   `json:"项目经历"`
	Skills     string         `json:"技能证书"`
	SelfEval   string         `json:"自我评价"`
}

// EducationExp 教育经历
type EducationExp struct {
	SchoolName  string `json:"学校名称"`
	Degree      string `json:"学历"`
	Major       string `json:"专业"`
	StartTime   string `json:"入学时间"`
	EndTime     string `json:"毕业时间"`
	Description string `json:"经历描述"`
}

// CampusExp 在校经历
type CampusExp struct {
	Title       string `json:"经历名称"`
	Role        string `json:"角色"`
	StartTime   string `json:"开始时间"`
	EndTime     string `json:"结束时间"`
	Description string `json:"经历描述"`
}

// InternExp 实习经历
type InternExp struct {
	Company     string `json:"公司名称"`
	Position    string `json:"职位"`
	StartTime   string `json:"开始时间"`
	EndTime     string `json:"结束时间"`
	Description string `json:"工作内容"`
}

// ProjectExp 项目经历
type ProjectExp struct {
	ProjectName string `json:"项目名称"`
	Role        string `json:"角色"`
	StartTime   string `json:"开始时间"`
	EndTime     string `json:"结束时间"`
	Description string `json:"项目描述"`
}

// RuleItem 评分规则项
type RuleItem struct {
	JudgmentDetail string  `json:"judgment_detail"` // 判定详情
	JudgmentScore  float64 `json:"judgment_score"`  // 判定分数
}

// ScoreRequest 评分请求
type ScoreRequest struct {
	Section map[string]interface{} `json:"section"` // 简历模块数据
	Rules   map[string][]RuleItem  `json:"rules"`   // 维度名 -> 判定项列表
}

// DimScore 维度得分
type DimScore struct {
	Rule  string  `json:"rule"`  // 维度名
	Score float64 `json:"score"` // 该维度得分
}

// ScoreResponse 评分响应（兼容多种返回格式）
type ScoreResponse struct {
	Code int         `json:"code,omitempty"`
	Msg  string      `json:"msg,omitempty"`
	Data []*DimScore `json:"data,omitempty"`
}

// AIWriteRequest AI 帮写请求
type AIWriteRequest struct {
	Resume      *ResumeData `json:"resume"`
	Requirement string      `json:"requirement"`
	Info        string      `json:"info"`
}
