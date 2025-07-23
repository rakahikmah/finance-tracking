package entity

import "time"

type Category struct {
	ID        int64     `gorm:"column:id"`
	CreatedBy int64     `gorm:"column:created_by"` // <-- Ini tetap exported agar GORM bisa memetakan
	Name      string    `gorm:"column:name"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (Category) TableName() string {
	return "categories"
}
