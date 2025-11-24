package schema

import (
	"cv2/internal/pkg/snowflake"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Position struct {
	ent.Schema
}

func (Position) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "rm_position",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_unicode_ci",
		},
		entsql.WithComments(true),
	}
}

func (Position) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Immutable().
			DefaultFunc(snowflake.NextID).
			Comment("id"),

		field.Int("level").
			Comment("层级"),

		field.Int64("parent_id").
			Default(0).
			Optional().
			Comment("父id"),

		field.String("title").
			NotEmpty().
			Comment("职位名称"),
	}
}

func (Position) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("children", Position.Type),
		edge.From("parent", Position.Type).
			Ref("children").
			Unique().
			Field("parent_id"),
	}
}
