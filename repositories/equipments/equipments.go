package equipments

import (
	"github.com/kaellybot/kaelly-encyclopedia/models/entities"
	"github.com/kaellybot/kaelly-encyclopedia/utils/databases"
)

func New(db databases.MySQLConnection) *Impl {
	return &Impl{db: db}
}

func (repo *Impl) GetEquipmentTypes() ([]entities.EquipmentType, error) {
	var equipmentTypes []entities.EquipmentType
	response := repo.db.GetDB().
		Model(&entities.EquipmentType{}).
		Preload("AreaEffects").
		Find(&equipmentTypes)
	return equipmentTypes, response.Error
}
