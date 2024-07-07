package user

import (
	"github.com/google/uuid"
	"github.com/isd-sgcu/rpkm67-model/model"
	"gorm.io/gorm"
)

type Repository interface {
	FindOne(id string, user *model.User) error
	AssignGroupTX(tx *gorm.DB, id string, groupID *uuid.UUID) error
}

type repositoryImpl struct {
	Db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repositoryImpl{Db: db}
}

func (r *repositoryImpl) FindOne(id string, user *model.User) error {
	return r.Db.Model(user).First(user, "id = ?", id).Error
}

func (r *repositoryImpl) AssignGroupTX(tx *gorm.DB, id string, groupID *uuid.UUID) error {
	return r.Db.Model(&model.User{}).Where("id = ?", id).Update("group_id", groupID).Error
}
