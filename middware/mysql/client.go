package mysql

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"time"
)

var DB *gorm.DB

// DBConfig 用于存储数据库配置
type DBConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// InitDB 初始化数据库连接
func InitDB(config DBConfig) {
	// 打开数据库连接
	var err error
	DB, err = gorm.Open(mysql.Open(config.DSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// 获取底层 *sql.DB
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("failed to get database instance: %v", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)       // 最大打开连接数
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)       // 最大空闲连接数
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime) // 最大连接生命周期
	sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime) // 最大空闲时间
}

// GetDBClient 返回数据库连接
func GetDBClient() *gorm.DB {
	return DB
}
