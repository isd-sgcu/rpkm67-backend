package group

import (
	"errors"

	"github.com/google/uuid"
	"github.com/isd-sgcu/rpkm67-model/model"
	"gorm.io/gorm"
)

type Repository interface {
	WithTransaction(txFunc func(*gorm.DB) error) error
	FindOne(id string, group *model.Group) error
	FindByToken(token string, group *model.Group) error
	UpdateConfirm(id string, group *model.Group) error
	CreateTX(tx *gorm.DB, group *model.Group) error
	DeleteGroupTX(tx *gorm.DB, groupId *uuid.UUID) error
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

func (r *repositoryImpl) FindOne(id string, group *model.Group) error {
	return r.Db.Preload("Members").First(&group, "id = ?", id).Error
}

func (r *repositoryImpl) FindByToken(token string, group *model.Group) error {
	return r.Db.Preload("Members").
		Joins("JOIN users ON users.id = groups.leader_id").
		First(&group, "token = ?", token).Error
}

func (r *repositoryImpl) UpdateConfirm(id string, group *model.Group) error {
	result := r.Db.Model(&model.Group{}).Where("id = ?", id).Update("is_confirmed", group.IsConfirmed)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no group found with the given id")
	}

	return nil
}

func (r *repositoryImpl) CreateTX(tx *gorm.DB, group *model.Group) error {
	return tx.Create(&group).Error
}

func (r *repositoryImpl) DeleteGroupTX(tx *gorm.DB, groupId *uuid.UUID) error {
	result := tx.Delete(&model.Group{}, "id = ?", groupId)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("not found group match with given id")
	}

	return nil
}
