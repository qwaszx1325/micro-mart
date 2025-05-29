package model

import (
	"time"

	"github.com/google/uuid"
	"micro-mart/services/transaction-orchestrator/domain/aggregate"
	"micro-mart/services/transaction-orchestrator/domain/entity"
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
	ID              string    `json:"id"`
	TransactionID   string    `json:"transaction_id"`
	ServiceName     string    `json:"service_name"`
	Operation       string    `json:"operation"`
	Status          string    `json:"status"`
	Payload         []byte    `json:"payload"`
	Result          []byte    `json:"result"`
	CompensationPayload []byte    `json:"compensation_payload"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
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

// ToAggregate converts a model.Transaction to an aggregate.Transaction
func (t *Transaction) ToAggregate() *aggregate.Transaction {
	steps := make([]entity.TransactionStep, len(t.Steps))
	for i, step := range t.Steps {
		steps[i] = entity.TransactionStep{
			ID:                  step.ID,
			TransactionID:       step.TransactionID,
			ServiceName:         step.ServiceName,
			Operation:           step.Operation,
			Status:              step.Status,
			Payload:             step.Payload,
			Result:              step.Result,
			CompensationPayload: step.CompensationPayload,
			CreatedAt:           step.CreatedAt,
			UpdatedAt:           step.UpdatedAt,
		}
	}

	return &aggregate.Transaction{
		ID:     uuid.MustParse(t.ID),
		Steps:  steps,
		Status: entity.TransactionStatus(t.Status),
	}
}

// FromAggregate converts an aggregate.Transaction to a model.Transaction
func FromAggregate(a *aggregate.Transaction) *Transaction {
	steps := make([]TransactionStep, len(a.Steps))
	for i, step := range a.Steps {
		steps[i] = TransactionStep{
			ID:                  step.ID,
			TransactionID:       step.TransactionID,
			ServiceName:         step.ServiceName,
			Operation:           step.Operation,
			Status:              step.Status,
			Payload:             step.Payload,
			Result:              step.Result,
			CompensationPayload: step.CompensationPayload,
			CreatedAt:           step.CreatedAt,
			UpdatedAt:           step.UpdatedAt,
		}
	}

	return &Transaction{
		ID:        a.ID.String(),
		Status:    TransactionStatus(a.Status),
		Steps:     steps,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}