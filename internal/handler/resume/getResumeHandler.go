// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package resume

import (
	"net/http"

	"cv2/internal/logic/resume"
	"cv2/internal/svc"
	"cv2/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取简历详情
func GetResumeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetResumeReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := resume.NewGetResumeLogic(r.Context(), svcCtx)
		resp, err := l.GetResume(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
