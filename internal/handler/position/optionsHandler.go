package position

import (
	"net/http"

	"cv2/internal/logic/position"
	"cv2/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// Position options
func OptionsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := position.NewOptionsLogic(r.Context(), svcCtx)
		resp, err := l.Options()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
