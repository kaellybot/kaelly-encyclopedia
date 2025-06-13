package encyclopedias

import (
	"context"
	"strconv"

	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-encyclopedia/models/constants"
	"github.com/kaellybot/kaelly-encyclopedia/models/mappers"
	"github.com/kaellybot/kaelly-encyclopedia/services/sources"
	"github.com/rs/zerolog/log"
)

func (service *Impl) itemRequest(ctx amqp.Context, message *amqp.RabbitMQMessage) {
	request := message.EncyclopediaItemRequest
	lg := mappers.MapLanguage(message.Language)
	if !isValidItemRequest(request) {
		service.replyWithFailedAnswer(ctx, amqp.RabbitMQMessage_ENCYCLOPEDIA_ITEM_ANSWER,
			message.Language)
		return
	}

	log.Info().Str(constants.LogCorrelationID, ctx.CorrelationID).
		Str(constants.LogQueryID, request.Query).
		Str(constants.LogQueryType, request.GetType().String()).
		Msgf("Encyclopedia Item Request received")

	funcs, found := service.getItemByFuncs[request.Type]
	if !found {
		log.Error().Str(constants.LogCorrelationID, ctx.CorrelationID).
			Str(constants.LogQueryID, request.Query).
			Str(constants.LogQueryType, request.GetType().String()).
			Msgf("Error while handling encyclopedia item query type, returning failed request")
		service.replyWithFailedAnswer(ctx, amqp.RabbitMQMessage_ENCYCLOPEDIA_ITEM_ANSWER,
			message.Language)
		return
	}

	var reply *amqp.EncyclopediaItemAnswer
	var err error
	if request.GetIsID() {
		ankamaID, errID := strconv.ParseInt(request.Query, 10, 32)
		if errID != nil {
			log.Error().Err(errID).
				Str(constants.LogCorrelationID, ctx.CorrelationID).
				Str(constants.LogQueryID, request.Query).
				Str(constants.LogQueryType, request.GetType().String()).
				Msgf("Error while converting query as ankamaID, returning failed request")
			service.replyWithFailedAnswer(ctx, amqp.RabbitMQMessage_ENCYCLOPEDIA_ITEM_ANSWER,
				message.Language)
			return
		}

		reply, err = funcs.GetItemByID(ctx, ankamaID, ctx.CorrelationID, lg)
	} else {
		reply, err = funcs.GetItemByQuery(ctx, request.Query, ctx.CorrelationID, lg)
	}

	if err != nil {
		log.Error().Err(err).
			Str(constants.LogCorrelationID, ctx.CorrelationID).
			Str(constants.LogQueryID, request.Query).
			Str(constants.LogQueryType, request.GetType().String()).
			Msgf("Error while retrieving encyclopedia item, returning failed request")
		service.replyWithFailedAnswer(ctx, amqp.RabbitMQMessage_ENCYCLOPEDIA_ITEM_ANSWER,
			message.Language)
		return
	}

	response := mappers.MapItem(reply, message.Language)
	service.replyWithSuceededAnswer(ctx, response)
}

func (service *Impl) getItemByID(_ context.Context, _ int64, _,
	_ string) (*amqp.EncyclopediaItemAnswer, error) {
	return nil, errBadRequestMessage
}

func (service *Impl) getItemByQuery(ctx context.Context, query, correlationID,
	lg string) (*amqp.EncyclopediaItemAnswer, error) {
	values, err := service.sourceService.SearchAnyItems(ctx, query, lg)
	if err != nil {
		return nil, err
	}

	if len(values) == 0 {
		return mappers.MapNoItem(query, amqp.ItemType_ANY_ITEM_TYPE), nil
	}

	// We trust the omnisearch by taking the first one in the list
	item := values[0]
	itemType := service.sourceService.GetItemType(item.Type.GetNameId())
	funcs, found := service.getItemByFuncs[itemType]
	if !found {
		return nil, sources.ErrFuncNotFound
	}

	resp, err := funcs.GetItemByID(ctx, int64(item.GetAnkamaId()), correlationID, lg)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func isValidItemRequest(request *amqp.EncyclopediaItemRequest) bool {
	return request != nil
}
