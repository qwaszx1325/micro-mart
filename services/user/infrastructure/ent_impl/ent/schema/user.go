// ent/schema/user.go
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"time"
)

// User 定義用戶實體
type User struct {
	ent.Schema
}

// Fields 定義用戶的欄位
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Comment("用戶UUID").
			Immutable(),
		field.String("username").
			Unique().
			Comment("用戶名"),
		field.String("password_hash").
			Sensitive().
			Comment("密碼雜湊"),
		field.String("email").
			Unique().
			Comment("電子郵件"),
		field.Time("created_at").
			Default(time.Now).
			Comment("創建時間"),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("更新時間"),
	}
}

// Edges 定義與其他實體的關係
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("roles", Role.Type).
			Comment("用戶擁有的角色"),
	}
}
