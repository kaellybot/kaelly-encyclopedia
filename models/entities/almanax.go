package entities

type Almanax struct {
	Day               int `gorm:"primaryKey"`
	Month             int `gorm:"primaryKey"`
	DofusDudeEffectID string
}
