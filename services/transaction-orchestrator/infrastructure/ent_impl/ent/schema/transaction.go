package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"time"
)

// Transaction defines the transaction entity
type Transaction struct {
	ent.Schema
}

// Fields defines the fields of the transaction
func (Transaction) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Comment("Transaction UUID").
			Immutable(),
		field.String("type").
			Comment("Transaction type"),
		field.Enum("status").
			Values("PENDING", "COMPLETED", "FAILED", "ROLLING_BACK", "ROLLED_BACK").
			Comment("Transaction status"),
		field.JSON("steps", []TransactionStep{}).
			Comment("Transaction steps"),
		field.Time("created_at").
			Default(time.Now).
			Comment("Creation time"),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("Update time"),
	}
}

// Edges defines the relations/edges of the transaction
func (Transaction) Edges() []ent.Edge {
	return []ent.Edge{
		// You can define relationships here if needed
	}
}

// TransactionStep represents a step in a transaction
type TransactionStep struct {
	ID                  string    `json:"id"`
	TransactionID       string    `json:"transaction_id"`
	ServiceName         string    `json:"service_name"`
	Operation           string    `json:"operation"`
	Status              string    `json:"status"`
	Payload             []byte    `json:"payload"`
	Result              []byte    `json:"result"`
	CompensationPayload []byte    `json:"compensation_payload"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}
