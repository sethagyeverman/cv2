// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package slot

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"cv2/internal/infra/ent"
	"cv2/internal/infra/ent/resumeslot"
	"cv2/internal/pkg/errx"
	"cv2/internal/pkg/sign"
	"cv2/internal/svc"
	"cv2/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SlotPayNotifyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 席位支付回调（供支付微服务调用）
func NewSlotPayNotifyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SlotPayNotifyLogic {
	return &SlotPayNotifyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SlotPayNotifyLogic) SlotPayNotify(req *types.SlotPayNotifyReq) (resp *types.SlotPayNotifyResp, err error) {
	// 1. 校验签名
	params := map[string]string{
		"order_id":     req.OrderID,
		"out_trade_no": req.OutTradeNo,
		"user_id":      req.UserID,
		"quantity":     strconv.Itoa(int(req.Quantity)),
		"status":       req.Status,
		"paid_at":      req.PaidAt,
	}

	if !sign.VerifyNotifySign(params, l.svcCtx.Config.Pay.NotifySecret, req.Sign) {
		l.Errorf("invalid signature: order_id=%s", req.OrderID)
		return &types.SlotPayNotifyResp{
			Success: false,
			Message: "invalid signature",
		}, nil
	}

	// 2. 校验支付状态
	if req.Status != "paid" {
		l.Infof("order not paid: order_id=%s, status=%s", req.OrderID, req.Status)
		return &types.SlotPayNotifyResp{
			Success: true,
			Message: "order not paid",
		}, nil
	}

	// 3. 幂等性检查：使用订单ID作为幂等键
	idempotentKey := fmt.Sprintf("slot:pay:notify:%s", req.OrderID)
	ok, err := l.svcCtx.Redis.SetNX(l.ctx, idempotentKey, "1", 24*time.Hour).Result()
	if err != nil {
		l.Errorf("redis setnx failed: %v", err)
		return &types.SlotPayNotifyResp{
			Success: false,
			Message: "idempotent check failed",
		}, nil
	}
	if !ok {
		l.Infof("duplicate notify: order_id=%s", req.OrderID)
		return &types.SlotPayNotifyResp{
			Success: true,
			Message: "duplicate notify, already processed",
		}, nil
	}

	// 4. 增加席位
	if err := l.addSlots(req.UserID, req.Quantity, req.OrderID); err != nil {
		// 增加席位失败，删除幂等键以便重试
		l.svcCtx.Redis.Del(l.ctx, idempotentKey)
		l.Errorf("add slots failed: %v", err)
		return &types.SlotPayNotifyResp{
			Success: false,
			Message: fmt.Sprintf("add slots failed: %v", err),
		}, nil
	}

	l.Infof("slot payment notify success: user_id=%s, quantity=%d, order_id=%s", req.UserID, req.Quantity, req.OrderID)
	return &types.SlotPayNotifyResp{
		Success: true,
		Message: "success",
	}, nil
}

// addSlots 增加用户席位（使用事务保证原子性）
func (l *SlotPayNotifyLogic) addSlots(userID string, quantity int32, orderID string) error {
	// 开启事务
	tx, err := l.svcCtx.Ent.Tx(l.ctx)
	if err != nil {
		return errx.Warp(http.StatusInternalServerError, err, "开启事务失败")
	}
	defer tx.Rollback()

	// 查询席位记录
	slot, err := tx.ResumeSlot.Query().
		Where(resumeslot.UserIDEQ(userID)).
		Only(l.ctx)

	if err != nil {
		// 记录不存在，创建新记录
		if ent.IsNotFound(err) {
			_, err = tx.ResumeSlot.Create().
				SetUserID(userID).
				SetMaxSlots(quantity).
				Save(l.ctx)
			if err != nil {
				return errx.Warp(http.StatusInternalServerError, err, "创建席位记录失败")
			}
		} else {
			return errx.Warp(http.StatusInternalServerError, err, "查询席位记录失败")
		}
	} else {
		// 记录存在，增加席位数
		_, err = tx.ResumeSlot.UpdateOne(slot).
			AddMaxSlots(quantity).
			Save(l.ctx)
		if err != nil {
			return errx.Warp(http.StatusInternalServerError, err, "更新席位数量失败")
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return errx.Warp(http.StatusInternalServerError, err, "提交事务失败")
	}

	l.Infof("successfully added %d slots for user %s, order %s", quantity, userID, orderID)
	return nil
}
