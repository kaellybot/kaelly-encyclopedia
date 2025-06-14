//nolint:lll // Clearer like that.
package entities

import amqp "github.com/kaellybot/kaelly-amqp"

type EquipmentType struct {
	EquipmentID amqp.EquipmentType `gorm:"primaryKey"`
	ItemID      amqp.ItemType      `gorm:"primaryKey"`
	DofusDudeID int32              `gorm:"primaryKey"`
	AreaEffects []WeaponAreaEffect `gorm:"many2many:equipment_type_weapon_area_effects;joinForeignKey:EquipmentID,ItemID,DofusDudeID;joinReferences:WeaponAreaEffectID"`
}

type WeaponAreaEffect struct {
	ID string `gorm:"primaryKey"`
}

type EquipmentTypeWeaponAreaEffect struct {
	EquipmentID        amqp.EquipmentType `gorm:"primaryKey"`
	ItemID             amqp.ItemType      `gorm:"primaryKey"`
	DofusDudeID        int32              `gorm:"primaryKey"`
	WeaponAreaEffectID string             `gorm:"primaryKey"`
	EquipmentType      EquipmentType      `gorm:"foreignKey:EquipmentID,ItemID,DofusDudeID;references:EquipmentID,ItemID,DofusDudeID;constraint:OnDelete:CASCADE"`
	WeaponAreaEffect   WeaponAreaEffect   `gorm:"foreignKey:WeaponAreaEffectID;references:ID;constraint:OnDelete:CASCADE"`
}

type WeaponException struct {
	DofusDudeID        int32            `gorm:"primaryKey"`
	WeaponAreaEffectID string           `gorm:"primaryKey"`
	WeaponAreaEffect   WeaponAreaEffect `gorm:"foreignKey:WeaponAreaEffectID;references:ID;constraint:OnDelete:CASCADE"`
}
