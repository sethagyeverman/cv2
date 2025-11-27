// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package resume

import (
	"net/http"

	"cv2/internal/logic/resume"
	"cv2/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 上传简历文件解析
func UploadResumeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 解析 multipart form
		if err := r.ParseMultipartForm(20 << 20); err != nil { // 20MB
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		file, fileHeader, err := r.FormFile("file")
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		defer file.Close()

		l := resume.NewUploadResumeLogic(r.Context(), svcCtx)
		resp, err := l.UploadResume(fileHeader)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
