package entity

import (
	"time"
)

// TransactionStatus represents the status of a transaction
type TransactionStatus string

const (
	StatusPending     TransactionStatus = "PENDING"
	StatusCompleted   TransactionStatus = "COMPLETED"
	StatusFailed      TransactionStatus = "FAILED"
	StatusRollingBack TransactionStatus = "ROLLING_BACK"
	StatusRolledBack  TransactionStatus = "ROLLED_BACK"
)

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

// Transaction represents a distributed transaction
type Transaction struct {
	ID        string            `json:"id"`
	Type      string            `json:"type"`
	Status    TransactionStatus `json:"status"`
	Steps     []TransactionStep `json:"steps"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}
