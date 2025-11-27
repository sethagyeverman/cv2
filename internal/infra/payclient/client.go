package payclient

import (
	"context"
	"time"
)

// 产品配置（硬编码）
const (
	SlotProductID   = "resume_slot"
	SlotProductName = "简历席位"
	SlotUnitPrice   = 100 // 单价（分）
)

// OrderStatus 订单状态
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"   // 待支付
	OrderStatusPaid      OrderStatus = "paid"      // 已支付
	OrderStatusExpired   OrderStatus = "expired"   // 已过期
	OrderStatusCancelled OrderStatus = "cancelled" // 已取消
)

// CreateOrderReq 创建订单请求
type CreateOrderReq struct {
	UserID      string `json:"user_id"`
	ProductID   string `json:"product_id"`
	ProductName string `json:"product_name"`
	Quantity    int32  `json:"quantity"`
	UnitPrice   int64  `json:"unit_price"`
	TotalAmount int64  `json:"total_amount"`
	NotifyURL   string `json:"notify_url"`
}

// CreateOrderResp 创建订单响应
type CreateOrderResp struct {
	OrderID     string    `json:"order_id"`
	CodeURL     string    `json:"code_url"`
	TotalAmount int64     `json:"total_amount"`
	ExpireTime  time.Time `json:"expire_time"`
}

// GetOrderResp 查询订单响应
type GetOrderResp struct {
	OrderID     string      `json:"order_id"`
	Status      OrderStatus `json:"status"`
	Quantity    int32       `json:"quantity"`
	TotalAmount int64       `json:"total_amount"`
	PaidAt      *time.Time  `json:"paid_at,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
}

// Client 支付客户端接口
type Client interface {
	// CreateOrder 创建订单
	CreateOrder(ctx context.Context, req *CreateOrderReq) (*CreateOrderResp, error)
	// GetOrder 查询订单
	GetOrder(ctx context.Context, orderID string) (*GetOrderResp, error)
}
