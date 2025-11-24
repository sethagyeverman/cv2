package resume

import (
	"net/http"

	"cv2/internal/logic/resume"
	"cv2/internal/svc"
	"cv2/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取简历上传 POST Policy
func GetUploadPolicyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UploadPolicyReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := resume.NewGetUploadPolicyLogic(r.Context(), svcCtx)
		resp, err := l.GetUploadPolicy(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
