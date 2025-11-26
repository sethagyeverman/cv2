// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package article

import (
	"context"

	"cv2/internal/svc"
	"cv2/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetArticleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取运营文章详情
func NewGetArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetArticleLogic {
	return &GetArticleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetArticleLogic) GetArticle(req *types.GetArticleReq) (resp *types.GetArticleResp, err error) {
	article, err := l.svcCtx.Shiji.GetArticle(l.ctx, req.ArticleId)
	if err != nil {
		l.Errorf("get article failed: article_id=%d, err=%v", req.ArticleId, err)
		return nil, err
	}

	return &types.GetArticleResp{
		Article: types.Article{
			ArticleId:    article.ArticleId,
			Title:        article.Title,
			Subtitle:     article.Subtitle,
			Content:      article.Content,
			ThumbnailUrl: article.ThumbnailUrl,
			Status:       article.Status,
			ViewCount:    article.ViewCount,
			PublishTime:  article.PublishTime,
			CreateTime:   article.CreateTime,
			UpdateTime:   article.UpdateTime,
		},
	}, nil
}
