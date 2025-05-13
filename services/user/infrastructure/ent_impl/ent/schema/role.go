// ent/schema/role.go
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Role 定義角色實體
type Role struct {
	ent.Schema
}

// Fields 定義角色的欄位
func (Role) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Comment("角色UUID").
			Immutable(),
		field.String("name").
			Unique().
			Comment("角色名稱"),
		field.String("description").
			Comment("角色描述"),
	}
}

// Edges 定義與其他實體的關係
func (Role) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("users", User.Type).
			Ref("roles").
			Comment("擁有此角色的用戶"),
		edge.To("permissions", Permission.Type).
			Comment("角色擁有的權限"),
	}
}
