// ent/schema/permission.go
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Permission 定義權限實體
type Permission struct {
	ent.Schema
}

// Fields 定義權限的欄位
func (Permission) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Comment("權限UUID").
			Immutable(),
		field.String("resource").
			Comment("資源名稱"),
		field.String("action").
			Comment("允許的操作：read, write, delete等"),
		field.Bool("allowed").
			Default(true).
			Comment("是否允許"),
	}
}

// Edges 定義與其他實體的關係
func (Permission) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("roles", Role.Type).
			Ref("permissions").
			Comment("擁有此權限的角色"),
	}
}
