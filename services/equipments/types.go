package equipments

import (
	"github.com/kaellybot/kaelly-encyclopedia/models/entities"
	repository "github.com/kaellybot/kaelly-encyclopedia/repositories/equipments"
	"github.com/kaellybot/kaelly-encyclopedia/repositories/weapons"
)

type Service interface {
	GetTypeByDofusDude(id int32) (entities.EquipmentType, bool)
	GetWeaponExceptions(id int32) []string
}

type Impl struct {
	dofusDudeTypes      map[int32]entities.EquipmentType
	weaponExceptions    map[int32][]string
	equipmentRepository repository.Repository
	weaponRepository    weapons.Repository
}
