package selection

import (
	"github.com/isd-sgcu/rpkm67-model/model"
	"gorm.io/gorm"
)

type Repository interface {
	Create(user *model.Selection) error
	FindByGroupId(groupId string, selections *[]model.Selection) error
	Delete(id string) error
	CountByBaanId() (map[string]int, error)
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
	return r.Db.Find(selections, "group_id = ?", groupId).Error
}

func (r *repositoryImpl) Delete(groupId string) error {
	return r.Db.Delete(&model.Selection{}, "group_id = ?", groupId).Error
}

func (r *repositoryImpl) CountByBaanId() (map[string]int, error) {
	var result []struct {
		Baan  string
		Count int
	}
	if err := r.Db.Model(&model.Selection{}).Select("baan, count(*) as count").Group("baan").Scan(&result).Error; err != nil {
		return nil, err
	}

	count := make(map[string]int)
	for _, v := range result {
		count[v.Baan] = v.Count
	}

	return count, nil
}
