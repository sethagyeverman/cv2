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

type ResumeScore struct {
	ent.Schema
}

func (ResumeScore) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "cv_resume_score",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_unicode_ci",
		},
		entsql.WithComments(true),
	}
}

func (ResumeScore) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Immutable().
			DefaultFunc(snowflake.NextID).
			Comment("得分ID"),

		field.Int64("resume_id").
			Comment("关联简历ID"),

		field.Int64("target_id").
			Comment("关联的模块ID或维度ID"),

		field.Int32("target_type").
			Comment("类型: 0=module, 1=dimension"),

		field.Float("score").
			Default(0).
			Comment("得分"),

		field.Float("weight").
			Default(0).
			Comment("权重"),

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

func (ResumeScore) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("resume", Resume.Type).
			Ref("scores").
			Unique().
			Required().
			Field("resume_id").
			Comment("关联简历"),
	}
}

func (ResumeScore) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resume_id", "target_type", "target_id"),
		index.Fields("target_type", "target_id"),
	}
}
