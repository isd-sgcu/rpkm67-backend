package group

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/isd-sgcu/rpkm67-model/model"
	"gorm.io/gorm"
)

type Repository interface {
	FindOne(userId string) (*model.Group, error)
}

type repositoryImpl struct {
	Db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repositoryImpl{
		Db: db,
	}
}

func (r *repositoryImpl) FindOne(userId string) (*model.Group, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	userUUID, err := uuid.Parse(userId)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID format: %v", err)
	}

	var group model.Group
	if err := r.Db.WithContext(ctx).
		Preload("Members").
		Joins("JOIN users ON users.group_id = groups.id").
		Where("users.id = ?", userUUID).
		First(&group).Error; err != nil {
		return nil, err
	}

	return &group, nil
}
