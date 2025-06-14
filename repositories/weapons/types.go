package weapons

import (
	"github.com/kaellybot/kaelly-encyclopedia/models/entities"
	"github.com/kaellybot/kaelly-encyclopedia/utils/databases"
)

type Repository interface {
	GetWeaponExceptions() ([]entities.WeaponException, error)
}

type Impl struct {
	db databases.MySQLConnection
}
