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
	FindOne(userId uuid.UUID) (*model.Group, error)
	FindByToken(token string) (*model.Group, error)
	Update(leaderUUID uuid.UUID, group *model.Group) error
}

type repositoryImpl struct {
	Db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repositoryImpl{
		Db: db,
	}
}

func (r *repositoryImpl) FindOne(userId uuid.UUID) (*model.Group, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var group model.Group
	if err := r.Db.WithContext(ctx).
		Preload("Members").
		Joins("JOIN users ON users.group_id = groups.id").
		Where("users.id = ?", userId).
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

func (r *repositoryImpl) Update(leaderUUID uuid.UUID, group *model.Group) error {
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
