package mappers

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/dofusdude/dodugo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-encyclopedia/models/constants"
	"github.com/kaellybot/kaelly-encyclopedia/services/equipments"
	"github.com/rs/zerolog/log"
)

func MapSetList(dodugoSets []dodugo.ListEquipmentSet) *amqp.EncyclopediaListAnswer {
	sets := make([]*amqp.EncyclopediaListAnswer_Item, 0)

	for _, set := range dodugoSets {
		sets = append(sets, &amqp.EncyclopediaListAnswer_Item{
			Id:   fmt.Sprintf("%v", set.GetAnkamaId()),
			Name: set.GetName(),
		})
	}

	return &amqp.EncyclopediaListAnswer{
		Items: sets,
	}
}

func MapSet(query string, set *dodugo.EquipmentSet, items map[int32]*dodugo.Weapon,
	icon string, equipmentService equipments.Service,
) *amqp.EncyclopediaItemAnswer {
	if set == nil {
		return &amqp.EncyclopediaItemAnswer{
			Type:   amqp.ItemType_SET_TYPE,
			Query:  query,
			Source: constants.GetDofusDudeSource(),
		}
	}

	equipments := make([]*amqp.EncyclopediaItemAnswer_Set_Equipment, 0)
	for _, itemID := range set.GetEquipmentIds() {
		formattedItemIDString := fmt.Sprintf("%v", itemID)
		item, found := items[itemID]
		if !found {
			log.Warn().
				Str(constants.LogAnkamaID, formattedItemIDString).
				Msgf("Cannot build entire set (missing item), continuing with degraded mode")
			missingItemID := itemID
			item = &dodugo.Weapon{
				AnkamaId: &missingItemID,
				Name:     &formattedItemIDString,
			}
		}

		equipments = append(equipments, &amqp.EncyclopediaItemAnswer_Set_Equipment{
			Id:    formattedItemIDString,
			Name:  item.GetName(),
			Level: int64(item.GetLevel()),
			Type:  mapEquipmentType(item.GetType(), equipmentService).EquipmentID,
		})
	}

	bonuses := make([]*amqp.EncyclopediaItemAnswer_Set_Bonus, 0)
	for itemNumberStr, bonus := range set.GetEffects() {
		// Ignore combination with no effects
		if len(bonus) == 0 {
			continue
		}

		itemNumber, err := strconv.ParseInt(itemNumberStr, 10, 64)
		if err != nil {
			log.Error().Err(err).
				Msgf("Cannot convert itemNumber '%v' as int64, ignoring this effect combination", itemNumberStr)
			continue
		}

		effects := make([]*amqp.EncyclopediaItemAnswer_Effect, 0)
		for _, effect := range bonus {
			effects = append(effects, &amqp.EncyclopediaItemAnswer_Effect{
				Id:    fmt.Sprintf("%v", *effect.GetType().Id),
				Label: effect.GetFormatted(),
			})
		}

		bonuses = append(bonuses, &amqp.EncyclopediaItemAnswer_Set_Bonus{
			ItemNumber: itemNumber,
			Effects:    effects,
		})
	}

	sort.Slice(bonuses, func(i, j int) bool {
		return bonuses[i].ItemNumber < bonuses[j].ItemNumber
	})

	return &amqp.EncyclopediaItemAnswer{
		Type:  amqp.ItemType_SET_TYPE,
		Query: query,
		Set: &amqp.EncyclopediaItemAnswer_Set{
			Id:         fmt.Sprintf("%v", set.GetAnkamaId()),
			Name:       set.GetName(),
			Level:      int64(set.GetHighestEquipmentLevel()),
			Icon:       icon,
			IsCosmetic: set.GetContainsCosmeticsOnly(),
			Equipments: equipments,
			Bonuses:    bonuses,
		},
		Source: constants.GetDofusDudeSource(),
	}
}
