package ent_impl

import (
	"context"
	"micro-mart/pkg/db"
	mmerror "micro-mart/pkg/mm_error"
	"micro-mart/pkg/mmotel"
	"micro-mart/services/user/domain/aggregate"
	"micro-mart/services/user/domain/entity"
	"micro-mart/services/user/domain/repository"
	"micro-mart/services/user/infrastructure/ent_impl/ent"
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
	ctx, span := mmotel.StartTrace(ctx)
	defer span.End()
	client := repo.db.GetClient(ctx).(*ent.Client)

	entUser, err := client.User.Create().
		SetEmail(u.Profile.Email).
		SetUsername(u.Profile.Name).
		SetPassword(u.Profile.Password).
		Save(ctx)

	if err != nil {
		return nil, mmerror.New(mmerror.InternalServerError, "register user fail")
	}
	user := &aggregate.User{
		ID: entUser.ID,
		Profile: entity.Profile{
			Email:    entUser.Email,
			Name:     entUser.Username,
			Password: entUser.Password,
		},
	}

	return user, nil
}

func (repo *UserRepository) GetUserByUserName(ctx context.Context, username string) (*aggregate.User, *mmerror.MmError) {

	return nil, nil
}
