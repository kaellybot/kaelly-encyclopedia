package sets

import (
	"github.com/kaellybot/kaelly-encyclopedia/models/entities"
	"github.com/kaellybot/kaelly-encyclopedia/utils/databases"
)

type Repository interface {
	GetSets() ([]entities.Set, error)
	Sync(newSets []entities.Set, unsyncSetIDs, deletedSetIDs []string) error
}

type Impl struct {
	db databases.MySQLConnection
}
