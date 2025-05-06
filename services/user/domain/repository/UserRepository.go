package repository

import (
	"context"
	mmerror "micro-mart/pkg/mm_error"
	"micro-mart/services/user/domain/aggregate"
)

type UserRepository interface {
	RegisterUser(ctx context.Context, u *aggregate.User) (*aggregate.User, *mmerror.MmError)
	GetUserByUserName(ctx context.Context, userName string) (*aggregate.User, *mmerror.MmError)
}
