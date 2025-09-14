package database

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Init 初始化数据库连接
func Init(dbPath string) (*gorm.DB, error) {
	// 配置 GORM 日志级别
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// 连接 SQLite 数据库
	db, err := gorm.Open(sqlite.Open(dbPath), config)
	if err != nil {
		return nil, err
	}

	// 设置连接池参数
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 设置最大打开连接数
	sqlDB.SetMaxOpenConns(10)
	// 设置最大空闲连接数
	sqlDB.SetMaxIdleConns(5)

	// 全局数据库实例
	DB = db

	log.Println("Database connected successfully")
	return db, nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}