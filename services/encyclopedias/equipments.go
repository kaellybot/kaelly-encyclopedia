package encyclopedias

import (
	"context"
	"errors"
	"fmt"

	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-encyclopedia/models/mappers"
	"github.com/kaellybot/kaelly-encyclopedia/services/sources"
)

func (service *Impl) getEquipmentByID(ctx context.Context, id int64, correlationID,
	lg string) (*amqp.EncyclopediaItemAnswer, error) {
	query := fmt.Sprintf("%v", id)
	equipment, err := service.sourceService.GetEquipmentByID(ctx, id, lg)
	if err != nil {
		if errors.Is(err, sources.ErrNotFound) {
			return mappers.MapEquipment(query, nil, nil, service.equipmentService), nil
		}

		return nil, err
	}

	ingredients := service.getIngredients(ctx, equipment.GetRecipe(), correlationID, lg)
	return mappers.MapEquipment(query, equipment, ingredients, service.equipmentService), nil
}

func (service *Impl) getEquipmentByQuery(ctx context.Context, query, correlationID,
	lg string) (*amqp.EncyclopediaItemAnswer, error) {
	equipment, err := service.sourceService.GetEquipmentByQuery(ctx, query, lg)
	if err != nil {
		if errors.Is(err, sources.ErrNotFound) {
			return mappers.MapEquipment(query, nil, nil, service.equipmentService), nil
		}

		return nil, err
	}

	ingredients := service.getIngredients(ctx, equipment.GetRecipe(), correlationID, lg)
	return mappers.MapEquipment(query, equipment, ingredients, service.equipmentService), nil
}
