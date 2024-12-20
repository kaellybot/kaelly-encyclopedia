package encyclopedias

import (
	"context"
	"fmt"

	"github.com/dofusdude/dodugo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-encyclopedia/models/constants"
	"github.com/kaellybot/kaelly-encyclopedia/models/mappers"
	"github.com/rs/zerolog/log"
)

func (service *Impl) getSetByID(ctx context.Context, id int64, correlationID,
	lg string) (*amqp.EncyclopediaItemAnswer, error) {
	set, err := service.sourceService.GetSetByID(ctx, id, lg)
	if err != nil {
		return nil, err
	}

	items := service.getSetEquipments(ctx, set, correlationID, lg)
	icon := service.getSetIcon(int64(set.GetAnkamaId()))
	return mappers.MapSet(set, items, icon, service.equipmentService), nil
}

func (service *Impl) getSetByQuery(ctx context.Context, query, correlationID,
	lg string) (*amqp.EncyclopediaItemAnswer, error) {
	set, err := service.sourceService.GetSetByQuery(ctx, query, lg)
	if err != nil {
		return nil, err
	}

	items := service.getSetEquipments(ctx, set, correlationID, lg)
	icon := service.getSetIcon(int64(set.GetAnkamaId()))
	return mappers.MapSet(set, items, icon, service.equipmentService), nil
}

func (service *Impl) getSetEquipments(ctx context.Context, set *dodugo.EquipmentSet, correlationID,
	lg string) map[int32]*dodugo.Weapon {
	var getItemByID func(ctx context.Context, equipmentID int64, lg string) (*dodugo.Weapon, error)
	if set.GetContainsCosmeticsOnly() {
		getItemByID = service.sourceService.GetCosmeticByID
	} else {
		getItemByID = service.sourceService.GetEquipmentByID
	}

	items := make(map[int32]*dodugo.Weapon)
	for _, itemID := range set.GetEquipmentIds() {
		item, errItem := getItemByID(ctx, int64(itemID), lg)
		if errItem != nil {
			log.Error().Err(errItem).
				Str(constants.LogCorrelationID, correlationID).
				Str(constants.LogAnkamaID, fmt.Sprintf("%v", itemID)).
				Msgf("Error while retrieving item with DofusDude, continuing without it")
		} else {
			items[itemID] = item
		}
	}

	return items
}

func (service *Impl) getSetIcon(setID int64) string {
	setDB, found := service.setService.GetSetByDofusDude(setID)
	if found {
		return setDB.Icon
	}

	return ""
}
