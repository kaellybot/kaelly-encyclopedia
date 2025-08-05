package sets

import (
	"github.com/kaellybot/kaelly-encyclopedia/models/entities"
	"github.com/kaellybot/kaelly-encyclopedia/utils/databases"
	"gorm.io/gorm"
)

func New(db databases.MySQLConnection) *Impl {
	return &Impl{db: db}
}

func (repo *Impl) GetSets() ([]entities.Set, error) {
	var sets []entities.Set
	response := repo.db.GetDB().
		Model(&entities.Set{}).
		Find(&sets)
	return sets, response.Error
}

func (repo *Impl) Sync(newSets []entities.Set, unsyncSetIDs, deletedSetIDs []string) error {
	return repo.db.GetDB().Transaction(func(tx *gorm.DB) error {
		if len(newSets) > 0 {
			if err := tx.Create(&newSets).Error; err != nil {
				return err
			}
		}

		if len(unsyncSetIDs) > 0 {
			if err := tx.Model(&entities.Set{}).
				Where("id IN ?", unsyncSetIDs).
				Update("is_current", false).Error; err != nil {
				return err
			}
		}

		if len(deletedSetIDs) > 0 {
			if err := tx.Where("dofus_dude_id IN ?", deletedSetIDs).Delete(&entities.Set{}).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
