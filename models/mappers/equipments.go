package mappers

import (
	"fmt"
	"slices"

	"github.com/dofusdude/dodugo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-encyclopedia/models/constants"
	"github.com/kaellybot/kaelly-encyclopedia/models/entities"
	"github.com/kaellybot/kaelly-encyclopedia/services/equipments"
	"github.com/rs/zerolog/log"
)

func MapEquipment(query string, item *dodugo.Weapon, ingredientItems map[int32]*constants.Ingredient,
	equipmentService equipments.Service) *amqp.EncyclopediaItemAnswer {
	if item == nil {
		return mapNilItem(query)
	}

	weaponEffects, effects := mapEffects(item.GetEffects())
	recipe := mapRecipe(item.GetRecipe(), ingredientItems)
	equipmentType := mapEquipmentType(item.GetType(), equipmentService)
	icon := item.GetImageUrls().Icon
	if item.GetImageUrls().Hq.IsSet() {
		icon = item.GetImageUrls().Hq.Get()
	}

	return &amqp.EncyclopediaItemAnswer{
		Type:  amqp.ItemType_EQUIPMENT_TYPE,
		Query: query,
		Equipment: &amqp.EncyclopediaItemAnswer_Equipment{
			Id:          fmt.Sprintf("%v", item.GetAnkamaId()),
			Name:        item.GetName(),
			Description: item.GetDescription(),
			Type: &amqp.EncyclopediaItemAnswer_Equipment_Type{
				ItemType:       equipmentType.ItemID,
				EquipmentType:  equipmentType.EquipmentID,
				EquipmentLabel: *item.GetType().Name,
			},
			Icon:            *icon,
			Level:           int64(item.GetLevel()),
			Pods:            int64(item.GetPods()),
			Set:             mapItemSet(item),
			Characteristics: mapCharacteristics(item, equipmentType, equipmentService),
			WeaponEffects:   weaponEffects,
			Effects:         effects,
			Conditions:      mapNullableConditions(item.Conditions),
			Recipe:          recipe,
		},
		Source: constants.GetDofusDudeSource(),
	}
}

func mapNilItem(query string) *amqp.EncyclopediaItemAnswer {
	return &amqp.EncyclopediaItemAnswer{
		Type:   amqp.ItemType_EQUIPMENT_TYPE,
		Query:  query,
		Source: constants.GetDofusDudeSource(),
	}
}

func mapItemSet(item *dodugo.Weapon) *amqp.EncyclopediaItemAnswer_Equipment_SetFamily {
	var set *amqp.EncyclopediaItemAnswer_Equipment_SetFamily
	if item.HasParentSet() {
		parentSet := item.GetParentSet()
		set = &amqp.EncyclopediaItemAnswer_Equipment_SetFamily{
			Id:   fmt.Sprintf("%v", parentSet.GetId()),
			Name: parentSet.GetName(),
		}
	}

	return set
}

func mapEffects(allEffects []dodugo.Effect,
) ([]*amqp.EncyclopediaItemAnswer_Effect, []*amqp.EncyclopediaItemAnswer_Effect) {
	weaponEffects := make([]*amqp.EncyclopediaItemAnswer_Effect, 0)
	effects := make([]*amqp.EncyclopediaItemAnswer_Effect, 0)
	for _, effect := range allEffects {
		amqpEffect := &amqp.EncyclopediaItemAnswer_Effect{
			Id:    fmt.Sprintf("%v", *effect.GetType().Id),
			Label: effect.GetFormatted(),
		}

		if effect.GetType().IsActive != nil && *effect.GetType().IsActive {
			weaponEffects = append(weaponEffects, amqpEffect)
		} else {
			effects = append(effects, amqpEffect)
		}
	}

	return weaponEffects, effects
}

func mapCharacteristics(item *dodugo.Weapon, itemType entities.EquipmentType, service equipments.Service,
) *amqp.EncyclopediaItemAnswer_Equipment_Characteristics {
	var characteristics *amqp.EncyclopediaItemAnswer_Equipment_Characteristics
	if item.GetIsWeapon() {
		areaEffectIDs := make([]string, 0)
		for _, effect := range itemType.AreaEffects {
			areaEffectIDs = append(areaEffectIDs, effect.ID)
		}

		particularEffectIDs := service.GetWeaponExceptions(item.GetAnkamaId())
		for _, effectID := range particularEffectIDs {
			if !slices.Contains(areaEffectIDs, effectID) {
				areaEffectIDs = append(areaEffectIDs, effectID)
			}
		}

		characteristics = &amqp.EncyclopediaItemAnswer_Equipment_Characteristics{
			Cost:           int64(item.GetApCost()),
			MinRange:       int64(item.Range.GetMin()),
			MaxRange:       int64(item.Range.GetMax()),
			MaxCastPerTurn: int64(item.GetMaxCastPerTurn()),
			CriticalRate:   int64(item.GetCriticalHitProbability()),
			CriticalBonus:  int64(item.GetCriticalHitBonus()),
			AreaEffectIds:  areaEffectIDs,
		}
	}

	return characteristics
}

func mapRecipe(recipe []dodugo.Recipe, ingredientItems map[int32]*constants.Ingredient,
) *amqp.EncyclopediaItemAnswer_Recipe {
	var response *amqp.EncyclopediaItemAnswer_Recipe
	if len(recipe) > 0 {
		ingredients := make([]*amqp.EncyclopediaItemAnswer_Recipe_Ingredient, 0)
		for _, recipeEntry := range recipe {
			formattedItemIDString := fmt.Sprintf("%v", recipeEntry.GetItemAnkamaId())
			ingredient, found := ingredientItems[recipeEntry.GetItemAnkamaId()]
			if !found {
				log.Warn().
					Str(constants.LogAnkamaID, formattedItemIDString).
					Msgf("Cannot build entire recipe (missing ingredient), continuing with degraded mode")
				ingredient = &constants.Ingredient{
					Name: formattedItemIDString,
					Type: amqp.ItemType_ANY_ITEM_TYPE,
				}
			}

			ingredients = append(ingredients, &amqp.EncyclopediaItemAnswer_Recipe_Ingredient{
				Id:       fmt.Sprintf("%v", recipeEntry.GetItemAnkamaId()),
				Name:     ingredient.Name,
				Quantity: int64(recipeEntry.GetQuantity()),
				Type:     ingredient.Type,
			})
		}

		response = &amqp.EncyclopediaItemAnswer_Recipe{
			Ingredients: ingredients,
		}
	}

	return response
}

func mapNullableConditions(conditions dodugo.NullableConditionNode,
) *amqp.EncyclopediaItemAnswer_Conditions {
	if !conditions.IsSet() {
		return nil
	}

	return mapConditions(conditions.Get())
}

func mapConditions(conditions *dodugo.ConditionNode,
) *amqp.EncyclopediaItemAnswer_Conditions {
	if conditions == nil {
		return nil
	}

	leaf := conditions.ConditionLeaf
	if leaf != nil {
		return &amqp.EncyclopediaItemAnswer_Conditions{
			Relation: amqp.EncyclopediaItemAnswer_Conditions_NONE,
			Condition: &amqp.EncyclopediaItemAnswer_Conditions_Condition{
				Operator: leaf.Condition.GetOperator(),
				Value:    int64(leaf.Condition.GetIntValue()),
				Element: &amqp.EncyclopediaItemAnswer_Conditions_Condition_Element{
					Id:   fmt.Sprintf("%v", leaf.Condition.Element.GetId()),
					Name: leaf.Condition.Element.GetName(),
				},
			},
		}
	}

	innerConditions := conditions.ConditionRelation
	if innerConditions != nil {
		var relation amqp.EncyclopediaItemAnswer_Conditions_Relation
		switch innerConditions.GetRelation() {
		case "or":
			relation = amqp.EncyclopediaItemAnswer_Conditions_OR
		case "and":
			relation = amqp.EncyclopediaItemAnswer_Conditions_AND
		default:
			log.Warn().
				Msgf("Cannot determine properly item condition relation '%v', using '%v' by default",
					innerConditions.GetRelation(), amqp.EncyclopediaItemAnswer_Conditions_NONE)
			relation = amqp.EncyclopediaItemAnswer_Conditions_NONE
		}

		children := make([]*amqp.EncyclopediaItemAnswer_Conditions, 0)
		for _, child := range innerConditions.GetChildren() {
			node := child
			children = append(children, mapConditions(node))
		}

		return &amqp.EncyclopediaItemAnswer_Conditions{
			Relation: relation,
			Children: children,
		}
	}

	return nil
}
