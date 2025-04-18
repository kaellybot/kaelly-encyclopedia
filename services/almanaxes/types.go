package almanaxes

import (
	"errors"
	"time"

	"github.com/kaellybot/kaelly-encyclopedia/models/entities"
	repository "github.com/kaellybot/kaelly-encyclopedia/repositories/almanaxes"
	"github.com/kaellybot/kaelly-encyclopedia/services/news"
	"github.com/kaellybot/kaelly-encyclopedia/services/sources"
)

var (
	errNotFound = errors.New("almanax is not found")
)

type Service interface {
	GetDatesByAlmanaxEffect(dofusDudeEffectID string) []time.Time
	GetLocation() *time.Location
}

type Impl struct {
	frenchLocation *time.Location
	almanaxes      map[string][]entities.Almanax
	sourceService  sources.Service
	newsService    news.Service
	repository     repository.Repository
}
