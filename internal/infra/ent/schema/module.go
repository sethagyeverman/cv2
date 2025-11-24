package schema

import (
	"cv2/internal/pkg/snowflake"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Module struct {
	ent.Schema
}

func (Module) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "cv_module",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_unicode_ci",
		},
		entsql.WithComments(true),
	}
}

func (Module) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Immutable().
			DefaultFunc(snowflake.NextID).
			Comment("模块ID"),

		field.String("title").
			NotEmpty().
			Comment("显示标题"),

		field.String("description").
			Default("").
			Comment("模块描述"),

		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("创建时间"),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("更新时间"),

		field.Time("deleted_at").
			Default(time.Time{}).
			Comment("删除时间"),
	}
}

func (Module) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("dimensions", Dimension.Type).
			Comment("模块下的维度"),
	}
}
