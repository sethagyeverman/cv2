package position

import (
	"context"
	"encoding/json"

	"cv2/internal/svc"
	"cv2/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type OptionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Position options
func NewOptionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OptionsLogic {
	return &OptionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OptionsLogic) Options() (resp *types.OptionsResp, err error) {
	type node struct {
		ID       int64   `json:"id"`
		Title    string  `json:"title"`
		Children []*node `json:"children,omitempty"`
	}

	positions, err := l.svcCtx.Ent.Position.Query().All(l.ctx)
	if err != nil {
		return nil, err
	}

	nodes := make(map[int64]*node)
	parent := make(map[int64]int64)
	var count int64

	for _, p := range positions {
		nodes[p.ID] = &node{ID: p.ID, Title: p.Title}
		parent[p.ID] = p.ParentID
		count++
	}

	var roots []*node
	for id, n := range nodes {
		pid := parent[id]
		if pid == 0 || nodes[pid] == nil {
			roots = append(roots, n)
			continue
		}
		nodes[pid].Children = append(nodes[pid].Children, n)
	}

	b, err := json.Marshal(roots)
	if err != nil {
		return nil, err
	}

	return &types.OptionsResp{Count: count, Data: string(b)}, nil
}
