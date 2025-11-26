package types

import "github.com/golang-jwt/jwt/v4"

// JWTClaims JWT 负载结构，包含用户信息
type JWTClaims struct {
	// 登录相关
	LoginType string `json:"loginType"` // 登录类型
	LoginID   string `json:"loginId"`   // 登录ID
	RnStr     string `json:"rnStr"`     // 随机字符串
	ClientID  string `json:"clientid"`  // 客户端ID

	// 用户信息
	UserID   int64  `json:"userId"`   // 用户ID
	UserName string `json:"userName"` // 用户名
	UserType string `json:"userType"` // 用户类型

	// 租户信息
	TenantID string `json:"tenantId"` // 租户ID

	// 部门信息
	DeptID       int64  `json:"deptId"`       // 部门ID
	DeptName     string `json:"deptName"`     // 部门名称
	DeptCategory string `json:"deptCategory"` // 部门分类

	// 过期时间相关
	ExpireTime      int64 `json:"expireTime,omitempty"`      // 过期时间戳
	LoginTime       int64 `json:"loginTime,omitempty"`       // 登录时间戳
	Timeout         int64 `json:"timeout,omitempty"`         // 超时时间
	TokenTimeout    int64 `json:"tokenTimeout,omitempty"`    // Token超时
	ActivityTimeout int64 `json:"activityTimeout,omitempty"` // 活动超时

	// JWT 标准字段
	jwt.RegisteredClaims
}

// UserContext 用户上下文信息（从 JWT 解析后存储到 context 中）
type UserContext struct {
	UserID   int64  `json:"user_id"`
	UserName string `json:"user_name"`
	TenantID string `json:"tenant_id"`
	UserType string `json:"user_type"`
	DeptID   int64  `json:"dept_id"`
	DeptName string `json:"dept_name"`
	LoginID  string `json:"login_id"`
	ClientID string `json:"client_id"`
}
