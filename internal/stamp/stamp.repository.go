package stamp

import (
	"github.com/isd-sgcu/rpkm67-model/model"
	"gorm.io/gorm"
)

type Repository interface {
	FindByUserId(userId string, stamp *model.Stamp) error
	StampByUserId(userId string, stamp *model.Stamp) error
	CreateAnswer(answer *model.Answer) error
}

type repositoryImpl struct {
	Db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repositoryImpl{
		Db: db,
	}
}

func (r *repositoryImpl) FindByUserId(userId string, stamp *model.Stamp) error {
	return r.Db.First(stamp, "user_id = ?", userId).Error
}

func (r *repositoryImpl) StampByUserId(userId string, stamp *model.Stamp) error {
	return r.Db.Model(stamp).Where("user_id = ?", userId).Updates(stamp).Error
}

func (r *repositoryImpl) CreateAnswer(answer *model.Answer) error {
	return r.Db.Create(answer).Error
}
