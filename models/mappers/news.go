package mappers

import (
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-encyclopedia/models/constants"
)

func MapAlmanaxNews(almanaxes []*amqp.NewsAlmanaxMessage_I18NAlmanax) *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_NEWS_ALMANAX,
		Language: amqp.Language_ANY,
		Game:     amqp.Game_DOFUS_GAME,
		NewsAlmanaxMessage: &amqp.NewsAlmanaxMessage{
			Almanaxes: almanaxes,
			Source:    constants.GetDofusDudeSource(),
		},
	}
}

func MapGameNews(gameVersion string) *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_NEWS_GAME,
		Language: amqp.Language_ANY,
		Game:     amqp.Game_DOFUS_GAME,
		NewsGameMessage: &amqp.NewsGameMessage{
			Version: gameVersion,
		},
	}
}

func MapSetNews(createdSetIDs, updatedSetIDs, deletedSetIDs []string) *amqp.RabbitMQMessage {
	return &amqp.RabbitMQMessage{
		Type:     amqp.RabbitMQMessage_NEWS_SET,
		Language: amqp.Language_ANY,
		Game:     amqp.Game_DOFUS_GAME,
		NewsSetMessage: &amqp.NewsSetMessage{
			CreatedSetIds: createdSetIDs,
			UpdatedSetIds: updatedSetIDs,
			DeletedSetIds: deletedSetIDs,
		},
	}
}
