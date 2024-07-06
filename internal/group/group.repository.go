package group

import (
	"errors"

	"github.com/google/uuid"
	"github.com/isd-sgcu/rpkm67-model/model"
	"gorm.io/gorm"
)

type Repository interface {
	WithTransaction(txFunc func(*gorm.DB) error) error
	FindByUserId(userId string, group *model.Group) error
	FindByToken(token string, group *model.Group) error
	Update(id string, group *model.Group) error
	CreateTX(tx *gorm.DB, group *model.Group) error
	MoveUserToNewGroupTX(tx *gorm.DB, userId string, groupId *uuid.UUID) error
	JoinGroupTX(tx *gorm.DB, userId string, groupId *uuid.UUID) error
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

func (r *repositoryImpl) FindByUserId(userId string, group *model.Group) error {
	return r.Db.Preload("Members").Preload("Selections").
		Where("id = (SELECT group_id FROM users WHERE id = ?)", userId).
		First(&group).Error
}

func (r *repositoryImpl) FindByToken(token string, group *model.Group) error {
	return r.Db.Preload("Members").
		Joins("JOIN users ON users.id = groups.leader_id").
		First(&group, "token = ?", token).Error
}

func (r *repositoryImpl) Update(id string, group *model.Group) error {
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

func (r *repositoryImpl) MoveUserToNewGroupTX(tx *gorm.DB, userId string, groupId *uuid.UUID) error {
	updateMap := map[string]interface{}{
		"group_id": groupId,
	}
	result := tx.Model(&model.User{}).Where("id = ?", userId).Updates(updateMap)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no user found with the given ID")
	}

	return nil
}

func (r *repositoryImpl) JoinGroupTX(tx *gorm.DB, userId string, groupId *uuid.UUID) error {
	updateMap := map[string]interface{}{
		"group_id": groupId,
	}

	result := tx.Model(&model.User{}).Where("id = ?", userId).Updates(updateMap)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no user found with the given ID")
	}

	return nil
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
