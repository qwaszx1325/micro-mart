package entimpl

import (
	"context"
	"micro-mart/pkg/db"
	mmerror "micro-mart/pkg/mm_error"
	"micro-mart/services/user/domain/aggregate"
	"micro-mart/services/user/domain/repository"
)

type UserRepository struct {
	db db.Database
}

var _ repository.UserRepository = (*UserRepository)(nil)

func NewUserRepository(db db.Database) repository.UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (repo *UserRepository) RegisterUser(ctx context.Context, u *aggregate.User) (*aggregate.User, *mmerror.MmError) {

	return nil, nil
}

func (repo *UserRepository) GetUserByUserName(ctx context.Context, username string) (*aggregate.User, *mmerror.MmError) {

	return nil, nil
}
