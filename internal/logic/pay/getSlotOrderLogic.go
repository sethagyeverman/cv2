// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package pay

import (
	"context"
	"fmt"

	"cv2/internal/svc"
	"cv2/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSlotOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 查询订单状态
func NewGetSlotOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSlotOrderLogic {
	return &GetSlotOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSlotOrderLogic) GetSlotOrder(req *types.GetSlotOrderReq) (resp *types.GetSlotOrderResp, err error) {
	// 1. 调用支付客户端查询订单
	order, err := l.svcCtx.PayClient.GetOrder(l.ctx, req.OrderID)
	if err != nil {
		l.Errorf("get order failed: %v", err)
		return nil, fmt.Errorf("get order failed: %w", err)
	}

	// 2. 构造响应
	resp = &types.GetSlotOrderResp{
		OrderID:     order.OrderID,
		Status:      string(order.Status),
		Quantity:    order.Quantity,
		TotalAmount: order.TotalAmount,
		CreatedAt:   order.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if order.PaidAt != nil {
		resp.PaidAt = order.PaidAt.Format("2006-01-02T15:04:05Z07:00")
	}

	return resp, nil
}
