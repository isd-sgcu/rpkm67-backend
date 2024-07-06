package group

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/isd-sgcu/rpkm67-model/model"
	"gorm.io/gorm"
)

type Repository interface {
	WithTransaction(txFunc func(*gorm.DB) error) error
	FindOne(userId *uuid.UUID) (*model.Group, error)
	FindByToken(token string) (*model.Group, error)
	Update(leaderUUID *uuid.UUID, group *model.Group) error
	DeleteMemberFromGroupWithTX(ctx context.Context, tx *gorm.DB, userUUID, groupUUID uuid.UUID) error
	CreateNewGroupWithTX(ctx context.Context, tx *gorm.DB, leaderId *uuid.UUID) (*model.Group, error)
	JoinGroupWithTX(ctx context.Context, tx *gorm.DB, userUUID, groupUUID uuid.UUID) error
	DeleteGroup(ctx context.Context, tx *gorm.DB, groupUUID uuid.UUID) error
}

type repositoryImpl struct {
	Db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repositoryImpl{
		Db: db,
	}
}

func (r *repositoryImpl) WithTransaction(txFunc func(*gorm.DB) error) error {
	tx := r.Db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := txFunc(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *repositoryImpl) FindOne(userId *uuid.UUID) (*model.Group, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var group model.Group
	if err := r.Db.WithContext(ctx).
		Preload("Members").
		Where("id = (SELECT group_id FROM users WHERE id = ?)", userId).
		First(&group).Error; err != nil {
		return nil, err
	}

	return &group, nil
}

func (r *repositoryImpl) FindByToken(token string) (*model.Group, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var group model.Group
	if err := r.Db.WithContext(ctx).
		Preload("Members").
		Joins("JOIN users ON users.id = groups.leader_id").
		First(&group, "token = ?", token).Error; err != nil {
		return nil, err
	}

	return &group, nil
}

func (r *repositoryImpl) Update(leaderUUID *uuid.UUID, group *model.Group) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	result := r.Db.WithContext(ctx).Model(&model.Group{}).Where("leader_id = ?", leaderUUID).Update("is_confirmed", group.IsConfirmed)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no promotion found with the given ID")
	}

	return nil
}

func (r *repositoryImpl) DeleteMemberFromGroupWithTX(ctx context.Context, tx *gorm.DB, userUUID, groupUUID uuid.UUID) error {
	updateMap := map[string]interface{}{
		"group_id": groupUUID,
	}
	result := r.Db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userUUID).Updates(updateMap)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no user found with the given ID")
	}

	return nil
}

func (r *repositoryImpl) CreateNewGroupWithTX(ctx context.Context, tx *gorm.DB, leaderId *uuid.UUID) (*model.Group, error) {
	group := model.Group{
		LeaderID: leaderId,
	}

	if err := r.Db.WithContext(ctx).Create(&group).Error; err != nil {
		return nil, err
	}

	return &group, nil
}

func (r *repositoryImpl) JoinGroupWithTX(ctx context.Context, tx *gorm.DB, userUUID, groupUUID uuid.UUID) error {
	updateMap := map[string]interface{}{
		"group_id": groupUUID,
	}

	result := r.Db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userUUID).Updates(updateMap)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no user found with the given ID")
	}

	return nil
}

func (r *repositoryImpl) DeleteGroup(ctx context.Context, tx *gorm.DB, groupUUID uuid.UUID) error {
	result := r.Db.Delete(&model.Group{}, "id = ?", groupUUID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("not found group match with given id")
	}

	return nil
}
