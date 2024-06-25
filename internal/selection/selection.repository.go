package selection

import (
	"github.com/isd-sgcu/rpkm67-auth/internal/model"
	"gorm.io/gorm"
)

type Repository interface {
	Create(user *model.Selection) error
	FindByGroupId(groupId string, selections *[]model.Selection) error
	Delete(id string) error
	CountGroupByBaanId() (map[string]int, error)
}

type repositoryImpl struct {
	Db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repositoryImpl{
		Db: db,
	}
}

func (r *repositoryImpl) Create(user *model.Selection) error {
	return r.Db.Create(user).Error
}

func (r *repositoryImpl) FindByGroupId(groupId string, selections *[]model.Selection) error {
	return r.Db.Find(selections, "groupId = ?", groupId).Error
}

func (r *repositoryImpl) Delete(id string) error {
	return r.Db.Delete(&model.Selection{}, "id = ?", id).Error
}

func (r *repositoryImpl) CountGroupByBaanId() (map[string]int, error) {
	var result []struct {
		BaanId string
		Count  int
	}
	if err := r.Db.Model(&model.Selection{}).Select("baan_id, count(*) as count").Group("baan_id").Scan(&result).Error; err != nil {
		return nil, err
	}

	count := make(map[string]int)
	for _, v := range result {
		count[v.BaanId] = v.Count
	}

	return count, nil
}
