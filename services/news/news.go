package news

import (
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-encyclopedia/models/mappers"
	"github.com/kaellybot/kaelly-encyclopedia/services/sources"
	"github.com/rs/zerolog/log"
)

func New(broker amqp.MessageBroker, sourceService sources.Service) *Impl {
	service := Impl{
		broker: broker,
	}

	sourceService.ListenGameEvent(service.PublishGameNews)
	return &service
}

func (service *Impl) PublishAlmanaxNews(almanaxes []*amqp.NewsAlmanaxMessage_I18NAlmanax) {
	log.Info().Msgf("Publishing almanax news...")
	err := service.broker.Emit(mappers.MapAlmanaxNews(almanaxes),
		amqp.ExchangeNews, newsAlmanaxRoutingKey, amqp.GenerateUUID())
	if err != nil {
		log.Error().Err(err).Msgf("Almanax news failed to be published")
	}
}

func (service *Impl) PublishGameNews(gameVersion string) {
	log.Info().Msgf("Publishing game version news...")
	err := service.broker.Emit(mappers.MapGameNews(gameVersion),
		amqp.ExchangeNews, newsGameRoutingKey, amqp.GenerateUUID())
	if err != nil {
		log.Error().Err(err).Msgf("Game news failed to be published")
	}
}

func (service *Impl) PublishSetNews(createdSetIDs, updatedSetIDs, deletedSetIDs []string) {
	log.Info().Msgf("Publishing sets check result...")
	err := service.broker.Emit(mappers.MapSetNews(createdSetIDs, updatedSetIDs, deletedSetIDs),
		amqp.ExchangeNews, newsSetRoutingKey, amqp.GenerateUUID())
	if err != nil {
		log.Error().Err(err).Msgf("Set news failed to be published")
	}
}
