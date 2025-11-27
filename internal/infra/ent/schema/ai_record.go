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

type AIRecord struct {
	ent.Schema
}

func (AIRecord) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "cv_ai_record",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_unicode_ci",
		},
		entsql.WithComments(true),
	}
}

func (AIRecord) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Immutable().
			DefaultFunc(snowflake.NextID).
			Comment("AI优化记录ID"),

		field.Int64("user_id").
			Comment("用户ID"),

		field.String("tenant_id").
			Comment("租户ID"),

		field.Int64("resume_id").
			Comment("简历ID"),

		field.Int64("module_id").
			Comment("模块ID"),

		field.Text("info").
			Optional().
			Default("").
			Comment("优化信息"),

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

func (AIRecord) Edges() []ent.Edge {
	return []ent.Edge{}
}

func (AIRecord) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "tenant_id"),
	}
}
