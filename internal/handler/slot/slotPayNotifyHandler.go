// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package slot

import (
	"net/http"

	"cv2/internal/logic/slot"
	"cv2/internal/svc"
	"cv2/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 席位支付回调（供支付微服务调用）
func SlotPayNotifyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SlotPayNotifyReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := slot.NewSlotPayNotifyLogic(r.Context(), svcCtx)
		resp, err := l.SlotPayNotify(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
