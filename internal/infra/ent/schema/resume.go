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

type Resume struct {
	ent.Schema
}

func (Resume) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "cv_resume",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_unicode_ci",
		},
		entsql.WithComments(true),
	}
}

func (Resume) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Immutable().
			DefaultFunc(snowflake.NextID).
			Comment("简历ID"),

		field.Int64("user_id").
			Comment("用户ID"),

		field.Int64("tenant_id").
			Comment("租户ID"),

		field.String("file_path").
			NotEmpty().
			Comment("文件路径"),

		field.String("file_name").
			NotEmpty().
			Comment("文件名"),

		field.Int32("status").
			Default(1).
			Comment("状态: 1=pending, 2=processing, 3=completed"),

		field.String("cover_image").
			Optional().
			Default("").
			Comment("封面图URL"),

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

func (Resume) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("scores", ResumeScore.Type).
			Comment("简历得分"),
	}
}

func (Resume) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "tenant_id"),
		index.Fields("status"),
	}
}
