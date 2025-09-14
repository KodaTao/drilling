package database

import (
	"log"

	"github.com/KodaTao/drilling/internal/models"
	"gorm.io/gorm"
)

// Migrate 执行数据库迁移
func Migrate(db *gorm.DB) error {
	log.Println("Running database migrations...")

	// 自动迁移数据库表结构
	err := db.AutoMigrate(&models.Host{}, &models.Tunnel{}, &models.ConnectionLog{}, &models.TrafficStats{})
	if err != nil {
		return err
	}

	log.Println("Database migrations completed successfully")
	return nil
}