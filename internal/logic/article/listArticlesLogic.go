// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package article

import (
	"context"

	"cv2/internal/svc"
	"cv2/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListArticlesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取运营文章列表
func NewListArticlesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListArticlesLogic {
	return &ListArticlesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListArticlesLogic) ListArticles(req *types.ListArticlesReq) (resp *types.ListArticlesResp, err error) {
	result, err := l.svcCtx.Shiji.ListArticles(l.ctx, req.PageNum, req.PageSize, req.Title)
	if err != nil {
		l.Errorf("list articles failed: %v", err)
		return nil, err
	}

	// 转换为响应类型
	articles := make([]types.Article, 0, len(result.Rows))
	for _, row := range result.Rows {
		articles = append(articles, types.Article{
			ArticleId:    row.ArticleId,
			Title:        row.Title,
			Subtitle:     row.Subtitle,
			Content:      row.Content,
			ThumbnailUrl: row.ThumbnailUrl,
			Status:       row.Status,
			ViewCount:    row.ViewCount,
			PublishTime:  row.PublishTime,
			CreateTime:   row.CreateTime,
			UpdateTime:   row.UpdateTime,
		})
	}

	return &types.ListArticlesResp{
		Rows:  articles,
		Total: result.Total,
	}, nil
}
