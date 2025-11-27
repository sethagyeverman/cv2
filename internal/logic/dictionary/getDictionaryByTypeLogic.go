package dictionary

import (
	"context"
	"net/http"

	"cv2/internal/infra/ent/dictionary"
	"cv2/internal/pkg/errx"
	"cv2/internal/svc"
	"cv2/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDictionaryByTypeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 根据类型获取字典
func NewGetDictionaryByTypeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDictionaryByTypeLogic {
	return &GetDictionaryByTypeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDictionaryByTypeLogic) GetDictionaryByType(req *types.GetDictionaryByTypeReq) (resp *types.GetDictionaryByTypeResp, err error) {
	// 验证参数
	if req.Type == "" {
		return nil, errx.New(http.StatusBadRequest, "字典类型不能为空")
	}

	// 查询指定类型的字典，按order排序
	items, err := l.svcCtx.Ent.Dictionary.Query().
		Where(dictionary.TypeEQ(req.Type)).
		Select(dictionary.FieldID, dictionary.FieldTitle).
		Order(dictionary.ByOrder()).
		All(l.ctx)
	if err != nil {
		return nil, errx.Warp(http.StatusInternalServerError, err, "查询字典失败")
	}

	// 转换为响应格式
	result := make([]types.DictionaryItem, 0, len(items))
	for _, item := range items {
		result = append(result, types.DictionaryItem{
			ID:    item.ID,
			Title: item.Title,
		})
	}

	l.Infof("get dictionary by type success, type=%s, count=%d", req.Type, len(result))
	return &types.GetDictionaryByTypeResp{Items: result}, nil
}
