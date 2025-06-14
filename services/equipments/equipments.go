package equipments

import (
	"github.com/kaellybot/kaelly-encyclopedia/models/constants"
	"github.com/kaellybot/kaelly-encyclopedia/models/entities"
	repository "github.com/kaellybot/kaelly-encyclopedia/repositories/equipments"
	"github.com/kaellybot/kaelly-encyclopedia/repositories/weapons"
	"github.com/rs/zerolog/log"
)

func New(repository repository.Repository, weaponRepository weapons.Repository,
) (*Impl, error) {
	equipmentTypes, errEquip := repository.GetEquipmentTypes()
	if errEquip != nil {
		return nil, errEquip
	}

	log.Info().
		Int(constants.LogEntityCount, len(equipmentTypes)).
		Msgf("Equipment types loaded")

	dofusDudeTypes := make(map[int32]entities.EquipmentType)
	for _, equipmentType := range equipmentTypes {
		dofusDudeTypes[equipmentType.DofusDudeID] = equipmentType
	}

	weaponExceptionRows, errWeapon := weaponRepository.GetWeaponExceptions()
	if errWeapon != nil {
		return nil, errWeapon
	}

	log.Info().
		Int(constants.LogEntityCount, len(weaponExceptionRows)).
		Msgf("Weapon exceptions loaded")

	weaponExceptions := make(map[int32][]string)
	for _, row := range weaponExceptionRows {
		exceptions := []string{row.WeaponAreaEffectID}
		if storedExceptions, found := weaponExceptions[row.DofusDudeID]; found {
			exceptions = append(storedExceptions, exceptions...)
		}
		weaponExceptions[row.DofusDudeID] = exceptions
	}

	return &Impl{
		dofusDudeTypes:      dofusDudeTypes,
		weaponExceptions:    weaponExceptions,
		equipmentRepository: repository,
		weaponRepository:    weaponRepository,
	}, nil
}

func (service *Impl) GetTypeByDofusDude(id int32) (entities.EquipmentType, bool) {
	item, found := service.dofusDudeTypes[id]
	return item, found
}

func (service *Impl) GetWeaponExceptions(id int32) []string {
	exceptions, found := service.weaponExceptions[id]
	if !found {
		return nil
	}

	return exceptions
}
