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
	"time"
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
	ctx, span := mmotel.StartSpan(ctx, "ApplicationUserService.Register")
	defer span.End()
	client := repo.db.GetClient(ctx).(*ent.Client)

	// 獲取當前時間用於設置創建時間和更新時間
	now := time.Now()

	// 創建用戶記錄
	entUser, err := client.User.Create().
		SetEmail(u.Profile.Email).
		SetUsername(u.Profile.Name).
		SetPasswordHash(u.Profile.Password). // 注意：實際應用中應使用雜湊後的密碼
		SetCreatedAt(now).                   // 明確設置創建時間
		SetUpdatedAt(now).                   // 明確設置更新時間
		Save(ctx)

	if err != nil {

		mmotel.Error(ctx, "Register user failed", mmotel.NewField("err", err))
		return nil, mmerror.New(mmerror.InternalServerError, "register user fail")
	}
	// 將數據庫實體轉換為領域模型
	user := &aggregate.User{
		ID: entUser.ID,
		Profile: entity.Profile{
			Email:     entUser.Email,
			Name:      entUser.Username,
			Password:  entUser.PasswordHash,
			CreatedAt: entUser.CreatedAt, // 將創建時間放入 Profile 中
			UpdatedAt: entUser.UpdatedAt, // 將更新時間放入 Profile 中
		},
	}

	return user, nil
}

func (repo *UserRepository) GetUserByUserName(ctx context.Context, username string) (*aggregate.User, *mmerror.MmError) {

	return nil, nil
}
