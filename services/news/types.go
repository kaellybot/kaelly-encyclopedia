package news

import (
	amqp "github.com/kaellybot/kaelly-amqp"
)

const (
	newsAlmanaxRoutingKey = "news.almanax"
	newsGameRoutingKey    = "news.game"
	newsSetRoutingKey     = "news.set"
)

type Service interface {
	PublishAlmanaxNews(almanaxes []*amqp.NewsAlmanaxMessage_I18NAlmanax)
	PublishGameNews(gameVersion string)
	PublishSetNews(createdSetIDs, updatedSetIDs, deletedSetIDs []string)
}

type Impl struct {
	broker amqp.MessageBroker
}
