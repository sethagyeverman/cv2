package payclient

import (
	"context"
	"fmt"
	"sync"
	"time"

	"cv2/internal/pkg/snowflake"
)

// MockClient 模拟支付客户端
type MockClient struct {
	mu     sync.RWMutex
	orders map[string]*mockOrder
}

type mockOrder struct {
	OrderID     string
	UserID      string
	Quantity    int32
	TotalAmount int64
	Status      OrderStatus
	PaidAt      *time.Time
	CreatedAt   time.Time
	ExpireTime  time.Time
}

// NewMockClient 创建模拟客户端
func NewMockClient() *MockClient {
	return &MockClient{
		orders: make(map[string]*mockOrder),
	}
}

// CreateOrder 创建订单（模拟）
func (c *MockClient) CreateOrder(ctx context.Context, req *CreateOrderReq) (*CreateOrderResp, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	orderID := fmt.Sprintf("%d", snowflake.NextID())
	now := time.Now()
	expireTime := now.Add(30 * time.Minute)

	order := &mockOrder{
		OrderID:     orderID,
		UserID:      req.UserID,
		Quantity:    req.Quantity,
		TotalAmount: req.TotalAmount,
		Status:      OrderStatusPending,
		CreatedAt:   now,
		ExpireTime:  expireTime,
	}
	c.orders[orderID] = order

	// 模拟微信支付二维码 URL
	codeURL := fmt.Sprintf("weixin://wxpay/bizpayurl?pr=mock_%s", orderID)

	return &CreateOrderResp{
		OrderID:     orderID,
		CodeURL:     codeURL,
		TotalAmount: req.TotalAmount,
		ExpireTime:  expireTime,
	}, nil
}

// GetOrder 查询订单（模拟）
func (c *MockClient) GetOrder(ctx context.Context, orderID string) (*GetOrderResp, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	order, ok := c.orders[orderID]
	if !ok {
		return nil, fmt.Errorf("order not found: %s", orderID)
	}

	// 检查是否过期
	status := order.Status
	if status == OrderStatusPending && time.Now().After(order.ExpireTime) {
		status = OrderStatusExpired
	}

	return &GetOrderResp{
		OrderID:     order.OrderID,
		Status:      status,
		Quantity:    order.Quantity,
		TotalAmount: order.TotalAmount,
		PaidAt:      order.PaidAt,
		CreatedAt:   order.CreatedAt,
	}, nil
}

// SimulatePay 模拟支付成功（仅用于测试）
func (c *MockClient) SimulatePay(orderID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	order, ok := c.orders[orderID]
	if !ok {
		return fmt.Errorf("order not found: %s", orderID)
	}

	if order.Status != OrderStatusPending {
		return fmt.Errorf("order status is not pending: %s", order.Status)
	}

	now := time.Now()
	order.Status = OrderStatusPaid
	order.PaidAt = &now

	return nil
}
