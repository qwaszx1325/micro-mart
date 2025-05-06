package aggregate

import (
	"github.com/google/uuid"
	"micro-mart/services/user/domain/entity"
)

type User struct {
	ID      uuid.UUID
	Profile entity.Profile
}
