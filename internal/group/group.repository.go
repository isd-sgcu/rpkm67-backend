package group

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/isd-sgcu/rpkm67-model/model"
	"gorm.io/gorm"
)

type Repository interface {
	FindOne(userId uuid.UUID) (*model.Group, error)
	FindByToken(token string) (*model.Group, error)
	Update(LeaderID uuid.UUID, group *model.Group) error
	Create(group *model.Group) error
	Join(user uuid.UUID, group *model.Group) error
	DeleteMember(userId uuid.UUID, group *model.Group) error
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
		Where("token = ?", token).
		First(&group).Error; err != nil {
		return nil, err
	}

	return &group, nil
}

func (r *repositoryImpl) Update(LeaderID uuid.UUID, group *model.Group) error {
	return r.Db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where(`leader_id = ?`, LeaderID).Update(`is_confirmed`, group.IsConfirmed).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *repositoryImpl) Create(group *model.Group) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := r.Db.WithContext(ctx).Create(group).Error; err != nil {
		return err
	}

	return nil
}

func (r *repositoryImpl) Join(userId uuid.UUID, group *model.Group) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	newMemberGroup, err := r.FindOne(userId)
	if err != nil {
		return err
	}

	var user model.User = *newMemberGroup.Members[0]
	if err := r.Db.WithContext(ctx).
		Model(&model.Group{}).
		Where("id = ?", group.ID).
		Association("Members").
		Append(user).Error; err != nil {
	}

	return nil
}

func (r *repositoryImpl) DeleteMember(userId uuid.UUID, group *model.Group) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := r.Db.WithContext(ctx).
		Preload("Members").
		Model(&model.Group{}).
		Where("id = ?", group.ID).
		Association("Members").
		Delete(&model.User{UUID: userId}).Error; err != nil {
	}

	return nil
}
