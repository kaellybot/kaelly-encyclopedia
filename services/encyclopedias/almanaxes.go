package encyclopedias

import (
	"context"
	"errors"

	"github.com/dofusdude/dodugo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-encyclopedia/models/constants"
	"github.com/kaellybot/kaelly-encyclopedia/models/mappers"
	"github.com/kaellybot/kaelly-encyclopedia/services/sources"
	"github.com/rs/zerolog/log"
)

func (service *Impl) almanaxRequest(ctx amqp.Context, message *amqp.RabbitMQMessage) {
	request := message.EncyclopediaAlmanaxRequest
	lg := mappers.MapLanguage(message.Language)
	if !isValidAlmanaxRequest(request) {
		service.replyWithFailedAnswer(ctx, amqp.RabbitMQMessage_ENCYCLOPEDIA_ALMANAX_ANSWER,
			message.Language)
		return
	}

	// Only case where we need to retrieve the precise local date to have the right day (and so the almanax).
	frenchDate := request.Date.AsTime().In(service.almanaxService.GetLocation())
	almanax, err := service.sourceService.GetAlmanaxByDate(ctx, frenchDate, lg)
	if err != nil {
		log.Error().Str(constants.LogCorrelationID, ctx.CorrelationID).
			Str(constants.LogDate, request.Date.String()).
			Msgf("Error while handling encyclopedia almanax date, returning failed request")
		service.replyWithFailedAnswer(ctx, amqp.RabbitMQMessage_ENCYCLOPEDIA_ALMANAX_ANSWER,
			message.Language)
		return
	}

	response := mappers.MapAlmanaxAnswer(almanax, service.sourceService, message.Language)
	service.replyWithSuceededAnswer(ctx, response)
}

func (service *Impl) almanaxEffectRequest(ctx amqp.Context, message *amqp.RabbitMQMessage) {
	request := message.EncyclopediaAlmanaxEffectRequest
	lg := mappers.MapLanguage(message.Language)
	if !isValidAlmanaxEffectRequest(request) {
		service.replyWithFailedAnswer(ctx, amqp.RabbitMQMessage_ENCYCLOPEDIA_ALMANAX_EFFECT_ANSWER,
			message.Language)
		return
	}

	log.Info().Str(constants.LogCorrelationID, ctx.CorrelationID).
		Msgf("Get almanax effect encyclopedia request received")

	effect, errEffect := service.getEffectFromRequest(ctx, request, lg)
	if errEffect != nil {
		if errors.Is(errEffect, sources.ErrNotFound) {
			response := mappers.MapAlmanaxEffects(request, "", nil, 0,
				service.sourceService, message.Language)
			service.replyWithSuceededAnswer(ctx, response)
			return
		}

		log.Error().Str(constants.LogCorrelationID, ctx.CorrelationID).
			Err(errEffect).
			Str(constants.LogQueryID, request.Query).
			Str(constants.LogDate, request.Date.String()).
			Msgf("Error while handling encyclopedia almanax effect request" +
				" and searching for accurate almanax effect, returning failed request")
		service.replyWithFailedAnswer(ctx, amqp.RabbitMQMessage_ENCYCLOPEDIA_ALMANAX_EFFECT_ANSWER,
			message.Language)
		return
	}

	offset := request.GetOffset()
	adjustedSize := offset + request.GetSize()
	dodugoAlmanaxes := make([]*dodugo.Almanax, 0)
	almanaxDates := service.almanaxService.GetDatesByAlmanaxEffect(*effect.Id)
	for i := offset; i < adjustedSize && i < int64(len(almanaxDates)); i++ {
		dodugoAlmanax, err := service.sourceService.GetAlmanaxByDate(ctx, almanaxDates[i], lg)
		if err != nil {
			log.Error().Str(constants.LogCorrelationID, ctx.CorrelationID).
				Str(constants.LogDate, almanaxDates[i].String()).
				Msgf("Error while handling encyclopedia almanax date, returning failed request")
			service.replyWithFailedAnswer(ctx, amqp.RabbitMQMessage_ENCYCLOPEDIA_ALMANAX_EFFECT_ANSWER,
				message.Language)
			return
		}

		dodugoAlmanaxes = append(dodugoAlmanaxes, dodugoAlmanax)
	}

	response := mappers.MapAlmanaxEffects(request, effect.GetName(), dodugoAlmanaxes,
		int64(len(almanaxDates)), service.sourceService, message.Language)
	service.replyWithSuceededAnswer(ctx, response)
}

func (service *Impl) almanaxResourceRequest(ctx amqp.Context, message *amqp.RabbitMQMessage) {
	request := message.EncyclopediaAlmanaxResourceRequest
	lg := mappers.MapLanguage(message.Language)
	if !isValidAlmanaxResourceRequest(request) {
		service.replyWithFailedAnswer(ctx, amqp.RabbitMQMessage_ENCYCLOPEDIA_ALMANAX_RESOURCE_ANSWER,
			message.Language)
		return
	}

	log.Info().Str(constants.LogCorrelationID, ctx.CorrelationID).
		Msgf("Get almanax resources encyclopedia request received")

	almanax, err := service.sourceService.GetAlmanaxByRange(ctx, request.Duration, lg)
	if err != nil {
		log.Error().Str(constants.LogCorrelationID, ctx.CorrelationID).
			Int64(constants.LogDuration, request.Duration).
			Msgf("Error while handling encyclopedia almanax resources, returning failed request")
		service.replyWithFailedAnswer(ctx, amqp.RabbitMQMessage_ENCYCLOPEDIA_ALMANAX_RESOURCE_ANSWER,
			message.Language)
		return
	}

	response := mappers.MapAlmanaxResource(almanax, request.Duration, service.sourceService, message.Language)
	service.replyWithSuceededAnswer(ctx, response)
}

func (service *Impl) getEffectFromRequest(ctx context.Context, request *amqp.EncyclopediaAlmanaxEffectRequest,
	lg string) (*dodugo.GetMetaAlmanaxBonuses200ResponseInner, error) {
	switch request.Type {
	case amqp.EncyclopediaAlmanaxEffectRequest_QUERY:
		values, errSearch := service.sourceService.SearchAlmanaxEffects(ctx, request.Query, lg)
		if errSearch != nil {
			return nil, errSearch
		}

		if len(values) == 0 {
			return nil, sources.ErrNotFound
		}

		// We trust the omnisearch by taking the first one in the list
		return &values[0], nil
	case amqp.EncyclopediaAlmanaxEffectRequest_DATE:
		dodugoAlmanax, errGet := service.sourceService.
			GetAlmanaxByDate(ctx, request.GetDate().AsTime(), lg)
		if errGet != nil {
			return nil, errGet
		}
		return dodugoAlmanax.Bonus.Type, nil
	default:
		return nil, errUnknownQuery
	}
}

func isValidAlmanaxRequest(request *amqp.EncyclopediaAlmanaxRequest) bool {
	return request != nil
}

func isValidAlmanaxEffectRequest(request *amqp.EncyclopediaAlmanaxEffectRequest) bool {
	return request != nil
}

func isValidAlmanaxResourceRequest(request *amqp.EncyclopediaAlmanaxResourceRequest) bool {
	return request != nil
}
