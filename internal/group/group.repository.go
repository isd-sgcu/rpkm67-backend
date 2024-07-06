package group

import (
	"errors"

	"github.com/google/uuid"
	"github.com/isd-sgcu/rpkm67-model/model"
	"gorm.io/gorm"
)

type Repository interface {
	WithTransaction(txFunc func(*gorm.DB) error) error
	FindOne(userId *uuid.UUID) (*model.Group, error)
	FindByToken(token string) (*model.Group, error)
	Update(leaderUUID *uuid.UUID, group *model.Group) error
	MoveUserToNewGroup(tx *gorm.DB, userUUID, groupUUID uuid.UUID) error
	CreateNewGroupWithTX(tx *gorm.DB, leaderId *uuid.UUID) (*model.Group, error)
	JoinGroupWithTX(tx *gorm.DB, userUUID, groupUUID uuid.UUID) error
	DeleteGroup(tx *gorm.DB, groupUUID uuid.UUID) error
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
	var group model.Group
	if err := r.Db.
		Preload("Members").
		Where("id = (SELECT group_id FROM users WHERE id = ?)", userId).
		First(&group).Error; err != nil {
		return nil, err
	}

	return &group, nil
}

func (r *repositoryImpl) FindByToken(token string) (*model.Group, error) {
	var group model.Group
	if err := r.Db.
		Preload("Members").
		Joins("JOIN users ON users.id = groups.leader_id").
		First(&group, "token = ?", token).Error; err != nil {
		return nil, err
	}

	return &group, nil
}

func (r *repositoryImpl) Update(leaderUUID *uuid.UUID, group *model.Group) error {
	result := r.Db.Model(&model.Group{}).Where("leader_id = ?", leaderUUID).Update("is_confirmed", group.IsConfirmed)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no promotion found with the given ID")
	}

	return nil
}

func (r *repositoryImpl) MoveUserToNewGroup(tx *gorm.DB, userUUID, groupUUID uuid.UUID) error {
	updateMap := map[string]interface{}{
		"group_id": groupUUID,
	}
	result := r.Db.Model(&model.User{}).Where("id = ?", userUUID).Updates(updateMap)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no user found with the given ID")
	}

	return nil
}

func (r *repositoryImpl) CreateNewGroupWithTX(tx *gorm.DB, leaderId *uuid.UUID) (*model.Group, error) {
	group := model.Group{
		LeaderID: leaderId,
	}

	if err := r.Db.Create(&group).Error; err != nil {
		return nil, err
	}

	return &group, nil
}

func (r *repositoryImpl) JoinGroupWithTX(tx *gorm.DB, userUUID, groupUUID uuid.UUID) error {
	updateMap := map[string]interface{}{
		"group_id": groupUUID,
	}

	result := r.Db.Model(&model.User{}).Where("id = ?", userUUID).Updates(updateMap)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no user found with the given ID")
	}

	return nil
}

func (r *repositoryImpl) DeleteGroup(tx *gorm.DB, groupUUID uuid.UUID) error {
	result := r.Db.Delete(&model.Group{}, "id = ?", groupUUID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("not found group match with given id")
	}

	return nil
}
