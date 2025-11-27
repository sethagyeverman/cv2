package schema

import (
	"cv2/internal/pkg/snowflake"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Dictionary struct {
	ent.Schema
}

func (Dictionary) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "rm_dictionary",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_unicode_ci",
		},
		entsql.WithComments(true),
	}
}

func (Dictionary) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Immutable().
			DefaultFunc(snowflake.NextID).
			Comment("字典ID"),

		field.String("title").
			NotEmpty().
			Comment("字典标题"),

		field.String("type").
			NotEmpty().
			Comment("字典类型"),

		field.Int("order").
			Default(0).
			Comment("排序顺序"),

		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("创建时间"),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("更新时间"),
	}
}

func (Dictionary) Edges() []ent.Edge {
	return []ent.Edge{}
}

func (Dictionary) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("type"),
	}
}
