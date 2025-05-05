package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable().
			Comment("用戶 UUID 主鍵"),
		field.String("username").
			Unique().
			NotEmpty(),
		field.String("email").
			Unique().
			NotEmpty(),
		field.String("password").NotEmpty().Sensitive(),
	}
}
