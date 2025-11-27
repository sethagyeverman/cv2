// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package slot

import (
	"context"
	"fmt"

	"cv2/internal/infra/payclient"
	"cv2/internal/svc"
	"cv2/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateSlotOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建席位购买订单
func NewCreateSlotOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateSlotOrderLogic {
	return &CreateSlotOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateSlotOrderLogic) CreateSlotOrder(req *types.CreateSlotOrderReq) (resp *types.CreateSlotOrderResp, err error) {
	// 1. 获取用户ID
	userID, ok := l.ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user_id not found in context")
	}

	// 2. 参数校验
	if req.Quantity <= 0 {
		return nil, fmt.Errorf("quantity must be greater than 0")
	}

	// 3. 构造支付客户端请求
	totalAmount := int64(req.Quantity) * payclient.SlotUnitPrice
	payReq := &payclient.CreateOrderReq{
		UserID:      userID,
		ProductID:   payclient.SlotProductID,
		ProductName: payclient.SlotProductName,
		Quantity:    req.Quantity,
		UnitPrice:   payclient.SlotUnitPrice,
		TotalAmount: totalAmount,
		NotifyURL:   l.svcCtx.Config.Pay.BuySlotNotifyURL,
	}

	// 4. 调用支付客户端创建订单
	payResp, err := l.svcCtx.PayClient.CreateOrder(l.ctx, payReq)
	if err != nil {
		l.Errorf("create order failed: %v", err)
		return nil, fmt.Errorf("create order failed: %w", err)
	}

	// 5. 返回响应
	return &types.CreateSlotOrderResp{
		OrderID:     payResp.OrderID,
		CodeURL:     payResp.CodeURL,
		TotalAmount: payResp.TotalAmount,
		ExpireTime:  payResp.ExpireTime.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
