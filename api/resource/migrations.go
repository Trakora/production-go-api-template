package resource

import (
	"production-go-api-template/api/resource/item"

	"gorm.io/gorm"
)

func AutoMigrateAll(db *gorm.DB) error {
	return db.AutoMigrate(
		&item.Item{},
	)
}
