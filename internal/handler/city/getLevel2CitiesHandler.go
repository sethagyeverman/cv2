package city

import (
	"net/http"

	"cv2/internal/logic/city"
	"cv2/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取所有二级城市
func GetLevel2CitiesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := city.NewGetLevel2CitiesLogic(r.Context(), svcCtx)
		resp, err := l.GetLevel2Cities()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
