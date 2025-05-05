package aggregate

import "micro-mart/services/user/domain/entity"

type User struct {
	ID      int64
	Profile entity.Profile
}
