package encyclopedias

import (
	"context"
	"fmt"

	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-encyclopedia/models/mappers"
)

func (service *Impl) getCosmeticByID(ctx context.Context, id int64, correlationID,
	lg string) (*amqp.EncyclopediaItemAnswer, error) {
	query := fmt.Sprintf("%v", id)
	cosmetic, err := service.sourceService.GetCosmeticByID(ctx, id, lg)
	if err != nil {
		return nil, err
	}

	ingredients := service.getIngredients(ctx, cosmetic.GetRecipe(), correlationID, lg)
	return mappers.MapEquipment(query, cosmetic, ingredients, service.equipmentService), nil
}

func (service *Impl) getCosmeticByQuery(ctx context.Context, query, correlationID,
	lg string) (*amqp.EncyclopediaItemAnswer, error) {
	cosmetic, err := service.sourceService.GetCosmeticByQuery(ctx, query, lg)
	if err != nil {
		return nil, err
	}

	ingredients := service.getIngredients(ctx, cosmetic.GetRecipe(), correlationID, lg)
	return mappers.MapEquipment(query, cosmetic, ingredients, service.equipmentService), nil
}
