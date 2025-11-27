package city

import (
	"context"
	"net/http"
	"strings"

	"cv2/internal/infra/ent/city"
	"cv2/internal/pkg/errx"
	"cv2/internal/svc"
	"cv2/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCitiesByInitialsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 根据首字母获取城市
func NewGetCitiesByInitialsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCitiesByInitialsLogic {
	return &GetCitiesByInitialsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCitiesByInitialsLogic) GetCitiesByInitials(req *types.GetCitiesByInitialsReq) (resp *types.GetCitiesByInitialsResp, err error) {
	// 将输入的首字母字符串转换为数组
	initials := strings.Split(strings.ToLower(req.Initials), "")
	if len(initials) == 0 {
		return nil, errx.New(http.StatusBadRequest, "首字母参数不能为空")
	}

	// 查询符合首字母的城市
	cities, err := l.svcCtx.Ent.City.Query().
		Where(city.InitialIn(initials...)).
		Select(city.FieldID, city.FieldTitle, city.FieldInitial).
		All(l.ctx)
	if err != nil {
		return nil, errx.Warp(http.StatusInternalServerError, err, "查询城市失败")
	}

	// 按首字母分组
	data := make(map[string][]types.CityItem)
	for _, c := range cities {
		initial := strings.ToLower(c.Initial)
		if initial == "" {
			continue
		}
		data[initial] = append(data[initial], types.CityItem{
			ID:    c.ID,
			Title: c.Title,
		})
	}

	l.Infof("get cities by initials success, initials=%s, count=%d", req.Initials, len(cities))
	return &types.GetCitiesByInitialsResp{Data: data}, nil
}
