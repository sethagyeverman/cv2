// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package auth

import (
	"context"
	"net/http"

	"cv2/internal/pkg/errx"
	"cv2/internal/svc"
	"cv2/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 用户登录
func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginReq) (resp *types.LoginResp, err error) {
	switch req.GrantType {
	case "password":
		return l.loginByPassword(req)
	default:
		return nil, errx.New(http.StatusBadRequest, "不支持的授权类型")
	}
}

// shiji 密码登录
func (l *LoginLogic) loginByPassword(req *types.LoginReq) (*types.LoginResp, error) {
	clientID := l.svcCtx.Config.OAuth2.Shiji.ClientID
	clientSecret := l.svcCtx.Config.OAuth2.Shiji.ClientSecret

	result, err := l.svcCtx.Shiji.Login(l.ctx, clientID, clientSecret, req.GrantType, req.Credentials)
	if err != nil {
		l.Errorf("password login failed: err=%v", err)
		return nil, err
	}

	l.Infof("password login success")
	return &types.LoginResp{
		AccessToken: result.AccessToken,
		ExpireIn:    result.ExpireIn,
	}, nil
}
