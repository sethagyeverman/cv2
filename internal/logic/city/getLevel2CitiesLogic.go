package city

import (
	"context"
	"net/http"

	"cv2/internal/infra/ent/city"
	"cv2/internal/pkg/errx"
	"cv2/internal/svc"
	"cv2/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLevel2CitiesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取所有二级城市
func NewGetLevel2CitiesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLevel2CitiesLogic {
	return &GetLevel2CitiesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetLevel2CitiesLogic) GetLevel2Cities() (resp *types.GetLevel2CitiesResp, err error) {
	// 查询所有二级城市（level=2）
	cities, err := l.svcCtx.Ent.City.Query().
		Where(city.Level(2)).
		Select(city.FieldID, city.FieldTitle).
		All(l.ctx)
	if err != nil {
		return nil, errx.Warp(http.StatusInternalServerError, err, "查询二级城市失败")
	}

	// 转换为响应格式
	items := make([]types.CityItem, 0, len(cities))
	for _, c := range cities {
		items = append(items, types.CityItem{
			ID:    c.ID,
			Title: c.Title,
		})
	}

	l.Infof("get level 2 cities success, count=%d", len(items))
	return &types.GetLevel2CitiesResp{Cities: items}, nil
}
