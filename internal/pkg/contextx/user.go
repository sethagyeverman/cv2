package contextx

import (
	"context"
	"fmt"
	"strconv"

	"cv2/internal/pkg/jwt"
	"cv2/internal/types"
)

// GetUserContext 从 context 中获取用户上下文
func GetUserContext(ctx context.Context) (*types.UserContext, error) {
	return jwt.GetUserContext(ctx)
}

// GetUserID 从 context 中获取用户ID
func GetUserID(ctx context.Context) (int64, error) {
	userCtx, err := GetUserContext(ctx)
	if err != nil {
		return 0, err
	}
	return userCtx.UserID, nil
}

// GetTenantID 从 context 中获取租户ID
func GetTenantID(ctx context.Context) (string, error) {
	userCtx, err := GetUserContext(ctx)
	if err != nil {
		return "", err
	}
	return userCtx.TenantID, nil
}

// GetTenantIDAsInt64 从 context 中获取租户ID（转换为 int64）
func GetTenantIDAsInt64(ctx context.Context) (int64, error) {
	tenantID, err := GetTenantID(ctx)
	if err != nil {
		return 0, err
	}

	id, err := strconv.ParseInt(tenantID, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid tenant_id format: %w", err)
	}
	return id, nil
}

// GetUserName 从 context 中获取用户名
func GetUserName(ctx context.Context) (string, error) {
	userCtx, err := GetUserContext(ctx)
	if err != nil {
		return "", err
	}
	return userCtx.UserName, nil
}

// GetDeptID 从 context 中获取部门ID
func GetDeptID(ctx context.Context) (int64, error) {
	userCtx, err := GetUserContext(ctx)
	if err != nil {
		return 0, err
	}
	return userCtx.DeptID, nil
}
