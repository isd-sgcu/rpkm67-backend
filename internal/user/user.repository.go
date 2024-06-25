package user

import (
	"github.com/isd-sgcu/rpkm67-auth/internal/model"
	"gorm.io/gorm"
)

type Repository interface {
	FindOne(id string, user *model.User) error
	FindByEmail(email string, user *model.User) error
	Create(user *model.User) error
	Update(id string, user *model.User) error
}

type repositoryImpl struct {
	Db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repositoryImpl{Db: db}
}

func (r *repositoryImpl) FindOne(id string, user *model.User) error {
	return r.Db.First(user, "id = ?", id).Error
}

func (r *repositoryImpl) FindByEmail(email string, user *model.User) error {
	return r.Db.First(user, "email = ?", email).Error
}

func (r *repositoryImpl) Create(user *model.User) error {
	return r.Db.Create(user).Error
}

func (r *repositoryImpl) Update(id string, user *model.User) error {
	return r.Db.Model(user).Where("id = ?", id).Updates(user).Error
}
