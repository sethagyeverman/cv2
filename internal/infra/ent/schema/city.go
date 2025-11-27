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

type City struct {
	ent.Schema
}

func (City) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "cv_city",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_unicode_ci",
		},
		entsql.WithComments(true),
	}
}

func (City) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Immutable().
			DefaultFunc(snowflake.NextID).
			Comment("城市ID"),

		field.Int64("parent_id").
			Default(0).
			Optional().
			Comment("父级城市ID"),

		field.Int("level").
			Comment("层级: 1=省份, 2=城市, 3=区县"),

		field.String("title").
			NotEmpty().
			Comment("城市名称"),

		field.String("initial").
			Optional().
			Default("").
			Comment("首字母"),

		field.Bool("is_hot").
			Default(false).
			Comment("是否热门城市"),

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

func (City) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("children", City.Type).
			Comment("子级城市"),
		edge.From("parent", City.Type).
			Ref("children").
			Unique().
			Field("parent_id").
			Comment("父级城市"),
	}
}

func (City) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("level"),
		index.Fields("initial"),
	}
}
