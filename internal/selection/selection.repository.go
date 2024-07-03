package selection

import (
	"github.com/isd-sgcu/rpkm67-model/model"
	"gorm.io/gorm"
)

type Repository interface {
	Create(user *model.Selection) error
	FindByGroupId(groupId string, selections *[]model.Selection) error
	Delete(groupId string) error
	CountByBaanId() (map[string]int, error)
	UpdateNewBaanExistOrder(updateSelection *model.Selection, selections *[]model.Selection) error
	UpdateExistBaanExistOrder(updateSelection *model.Selection, selections *[]model.Selection) error
	UpdateExistBaanNewOrder(updateSelection *model.Selection, selections *[]model.Selection) error
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

func (r *repositoryImpl) UpdateNewBaanExistOrder(updateSelection *model.Selection, selections *[]model.Selection) error {
	return r.Db.Transaction(func(tx *gorm.DB) error {
		var existingSelection model.Selection
		if err := tx.Where(`group_id = ? AND "order" = ?`, updateSelection.GroupID, updateSelection.Order).First(&existingSelection).Error; err != nil {
			return err
		}

		if err := tx.Where(`"order" = ?`, updateSelection.Order).Model(&existingSelection).Update("baan", updateSelection.Baan).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *repositoryImpl) UpdateExistBaanExistOrder(updateSelection *model.Selection, selections *[]model.Selection) error {
	return r.Db.Transaction(func(tx *gorm.DB) error {
		var existingBaanSelection model.Selection
		if err := tx.Where("group_id = ? AND baan = ?", updateSelection.GroupID, updateSelection.Baan).First(&existingBaanSelection).Error; err != nil {
			return err
		}

		var existingOrderSelection model.Selection
		if err := tx.Where(`group_id = ? AND "order" = ?`, updateSelection.GroupID, updateSelection.Order).First(&existingOrderSelection).Error; err != nil {
			return err
		}

		if existingBaanSelection.Order == updateSelection.Order {
			return nil
		}

		if err := tx.Where(`"order" = ?`, existingBaanSelection.Order).Model(&existingBaanSelection).Update("baan", existingOrderSelection.Baan).Error; err != nil {
			return err
		}
		if err := tx.Where(`"order" = ?`, existingOrderSelection.Order).Model(&existingOrderSelection).Update("baan", existingBaanSelection.Baan).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *repositoryImpl) UpdateExistBaanNewOrder(updateSelection *model.Selection, selections *[]model.Selection) error {
	return r.Db.Transaction(func(tx *gorm.DB) error {
		var existingSelection model.Selection
		if err := tx.Where("group_id = ? AND baan = ?", updateSelection.GroupID, updateSelection.Baan).First(&existingSelection).Error; err != nil {
			return err
		}

		if err := tx.Where("baan = ?", updateSelection.Baan).Model(&existingSelection).Update("order", updateSelection.Order).Error; err != nil {
			return err
		}

		return nil
	})
}
