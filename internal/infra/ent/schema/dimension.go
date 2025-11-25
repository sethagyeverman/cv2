package schema

import (
	"cv2/internal/pkg/snowflake"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Dimension struct {
	ent.Schema
}

func (Dimension) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "cv_dimension",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_unicode_ci",
		},
		entsql.WithComments(true),
	}
}

func (Dimension) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Immutable().
			DefaultFunc(snowflake.NextID).
			Comment("维度ID"),

		field.Int64("module_id").
			Comment("所属模块ID"),

		field.String("description").
			Comment("描述"),

		field.String("title").
			NotEmpty().
			Comment("显示标题"),

		field.JSON("judgment", []map[string]interface{}{}).
			Optional().
			Comment("维度得分详情（包括detail, score, weight的数组）"),

		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("创建时间"),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("更新时间"),

		field.Time("deleted_at").
			Optional().
			Nillable().
			Comment("删除时间"),
	}
}

func (Dimension) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("module", Module.Type).
			Ref("dimensions").
			Unique().
			Required().
			Field("module_id").
			Comment("所属模块"),
	}
}

func (Dimension) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("module_id"),
	}
}
