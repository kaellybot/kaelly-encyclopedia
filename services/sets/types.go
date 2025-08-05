package sets

import (
	"sync"

	"github.com/kaellybot/kaelly-encyclopedia/models/entities"
	repository "github.com/kaellybot/kaelly-encyclopedia/repositories/sets"
	"github.com/kaellybot/kaelly-encyclopedia/services/equipments"
	"github.com/kaellybot/kaelly-encyclopedia/services/news"
	"github.com/kaellybot/kaelly-encyclopedia/services/sources"
)

type Service interface {
	GetSetByDofusDude(ID int64) (entities.Set, bool)
}

type Impl struct {
	mu               sync.RWMutex
	sets             map[int64]entities.Set
	newsService      news.Service
	sourceService    sources.Service
	equipmentService equipments.Service
	repository       repository.Repository
}
