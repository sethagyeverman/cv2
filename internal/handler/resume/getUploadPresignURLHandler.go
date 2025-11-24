package resume

import (
	"net/http"

	"cv2/internal/logic/resume"
	"cv2/internal/svc"
	"cv2/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取简历上传预签名 URL
func GetUploadPresignURLHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UploadPresignReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := resume.NewGetUploadPresignURLLogic(r.Context(), svcCtx)
		resp, err := l.GetUploadPresignURL(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
