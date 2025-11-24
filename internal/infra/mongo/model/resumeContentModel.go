package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ResumeContent MongoDB 中存储的简历内容
type ResumeContent struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	MySQLID    int64              `bson:"mysql_id" json:"mysql_id"`       // 关联 MySQL 中的简历ID
	CreateTime time.Time          `bson:"create_time" json:"create_time"` // 创建时间
	UpdateTime time.Time          `bson:"update_time" json:"update_time"` // 更新时间
	RawMD      string             `bson:"raw_md" json:"raw_md"`           // 原始 Markdown 内容
	Modules    []ModuleData       `bson:"modules" json:"modules"`         // 模块化数据
}

// ModuleData 模块数据
type ModuleData struct {
	ModuleID int64                    `bson:"module_id" json:"module_id"` // 模块ID，对应 MySQL 中的 module.id
	Title    string                   `bson:"title" json:"title"`         // 模块标题
	Data     []map[string]interface{} `bson:"data" json:"data"`           // 模块具体数据（灵活结构）
}

// BasicInfo 基本信息模块
type BasicInfo struct {
	Name      string      `bson:"name" json:"name"`
	Gender    string      `bson:"gender,omitempty" json:"gender,omitempty"`
	BirthDate *time.Time  `bson:"birth_date,omitempty" json:"birth_date,omitempty"`
	Contact   ContactInfo `bson:"contact,omitempty" json:"contact,omitempty"`
}

// ContactInfo 联系方式
type ContactInfo struct {
	Phone string `bson:"phone,omitempty" json:"phone,omitempty"`
	Email string `bson:"email,omitempty" json:"email,omitempty"`
}

// Education 教育背景
type Education struct {
	Degree      string     `bson:"degree" json:"degree"`
	Major       string     `bson:"major" json:"major"`
	School      string     `bson:"school" json:"school"`
	StartDate   *time.Time `bson:"start_date,omitempty" json:"start_date,omitempty"`
	EndDate     *time.Time `bson:"end_date,omitempty" json:"end_date,omitempty"`
	Description string     `bson:"description,omitempty" json:"description,omitempty"`
}

// Project 项目经历
type Project struct {
	Name         string     `bson:"name" json:"name"`
	Role         string     `bson:"role,omitempty" json:"role,omitempty"`
	Description  string     `bson:"description,omitempty" json:"description,omitempty"`
	StartDate    *time.Time `bson:"start_date,omitempty" json:"start_date,omitempty"`
	EndDate      *time.Time `bson:"end_date,omitempty" json:"end_date,omitempty"`
	Technologies []string   `bson:"technologies,omitempty" json:"technologies,omitempty"`
}

// Internship 实习经历
type Internship struct {
	Company     string     `bson:"company" json:"company"`
	Position    string     `bson:"position" json:"position"`
	StartDate   *time.Time `bson:"start_date,omitempty" json:"start_date,omitempty"`
	EndDate     *time.Time `bson:"end_date,omitempty" json:"end_date,omitempty"`
	Achievement string     `bson:"achievement,omitempty" json:"achievement,omitempty"`
}

// SchoolActivity 在校经历
type SchoolActivity struct {
	Name        string     `bson:"name" json:"name"`
	Role        string     `bson:"role,omitempty" json:"role,omitempty"`
	Description string     `bson:"description,omitempty" json:"description,omitempty"`
	StartDate   *time.Time `bson:"start_date,omitempty" json:"start_date,omitempty"`
	EndDate     *time.Time `bson:"end_date,omitempty" json:"end_date,omitempty"`
}

// Skill 技能证书
type Skill struct {
	Name        string `bson:"name" json:"name"`
	Level       string `bson:"level,omitempty" json:"level,omitempty"`
	Certificate string `bson:"certificate,omitempty" json:"certificate,omitempty"`
}

// SelfEvaluation 自我评价
type SelfEvaluation struct {
	Content    string   `bson:"content" json:"content"`
	Advantages []string `bson:"advantages,omitempty" json:"advantages,omitempty"`
}
