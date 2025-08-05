package entities

import amqp "github.com/kaellybot/kaelly-amqp"

type Set struct {
	ID        int32     `gorm:"primaryKey"`
	Game      amqp.Game `gorm:"primaryKey"`
	Hash      string
	Icon      string
	IsCurrent bool
}
