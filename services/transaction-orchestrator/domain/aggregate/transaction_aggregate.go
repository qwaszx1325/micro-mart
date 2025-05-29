package aggregate

import (
	"github.com/google/uuid"
	"micro-mart/services/transaction-orchestrator/domain/entity"
)

type Transaction struct {
	ID     uuid.UUID
	Steps  []entity.TransactionStep
	Status entity.TransactionStatus
}
