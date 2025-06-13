package sources

import (
	"context"
	"errors"
	"time"

	"github.com/dofusdude/dodugo"
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-encyclopedia/repositories/games"
	"github.com/kaellybot/kaelly-encyclopedia/services/stores"
)

type objectType string

const (
	almanax       objectType = "almanax"
	almanaxRange  objectType = "almanaxRange"
	almanaxEffect objectType = "almanaxEffect"
	item          objectType = "items"
	set           objectType = "sets"
)

var (
	ErrFuncNotFound = errors.New("no possibility to retrieve item")
	ErrNotFound     = errors.New("cannot find the desired resource")
)

type GameEventHandler func(gameVersion string)

type Service interface {
	GetItemType(itemType string) amqp.ItemType

	SearchAnyItems(ctx context.Context, query, lg string) ([]dodugo.GameSearch, error)
	SearchCosmetics(ctx context.Context, query, lg string) ([]dodugo.ListItem, error)
	SearchEquipments(ctx context.Context, query, lg string) ([]dodugo.ListItem, error)
	SearchMounts(ctx context.Context, query, lg string) ([]dodugo.Mount, error)
	SearchSets(ctx context.Context, query, lg string) ([]dodugo.ListEquipmentSet, error)
	SearchAlmanaxEffects(ctx context.Context, query, lg string) ([]dodugo.GetMetaAlmanaxBonuses200ResponseInner, error)

	GetConsumableByID(ctx context.Context, consumableID int64, lg string) (*dodugo.Resource, error)
	GetCosmeticByID(ctx context.Context, cosmeticID int64, lg string) (*dodugo.Weapon, error)
	GetEquipmentByID(ctx context.Context, equipmentID int64, lg string) (*dodugo.Weapon, error)
	GetMountByID(ctx context.Context, mountID int64, lg string) (*dodugo.Mount, error)
	GetQuestItemByID(ctx context.Context, questItemID int64, lg string) (*dodugo.Resource, error)
	GetResourceByID(ctx context.Context, resourceID int64, lg string) (*dodugo.Resource, error)
	GetSetByID(ctx context.Context, setID int64, lg string) (*dodugo.EquipmentSet, error)
	GetSets(ctx context.Context) ([]dodugo.ListEquipmentSet, error)

	GetCosmeticByQuery(ctx context.Context, query, lg string) (*dodugo.Weapon, error)
	GetEquipmentByQuery(ctx context.Context, query, lg string) (*dodugo.Weapon, error)
	GetMountByQuery(ctx context.Context, query, lg string) (*dodugo.Mount, error)
	GetSetByQuery(ctx context.Context, query, lg string) (*dodugo.EquipmentSet, error)

	GetAlmanaxByDate(ctx context.Context, date time.Time, language string) (*dodugo.Almanax, error)
	GetAlmanaxByRange(ctx context.Context, daysDuration int64, language string) ([]dodugo.Almanax, error)

	ListenGameEvent(handler GameEventHandler)
}

type Impl struct {
	eventHandlers   []GameEventHandler
	dofusDudeClient *dodugo.APIClient
	storeService    stores.Service
	gameRepo        games.Repository
	httpTimeout     time.Duration
	itemTypes       map[string]amqp.ItemType
}
