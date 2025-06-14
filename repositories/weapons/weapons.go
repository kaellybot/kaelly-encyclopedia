package weapons

import (
	"github.com/kaellybot/kaelly-encyclopedia/models/entities"
	"github.com/kaellybot/kaelly-encyclopedia/utils/databases"
)

func New(db databases.MySQLConnection) *Impl {
	return &Impl{db: db}
}

func (repo *Impl) GetWeaponExceptions() ([]entities.WeaponException, error) {
	var weaponExceptions []entities.WeaponException
	response := repo.db.GetDB().
		Model(&entities.WeaponException{}).
		Find(&weaponExceptions)
	return weaponExceptions, response.Error
}
