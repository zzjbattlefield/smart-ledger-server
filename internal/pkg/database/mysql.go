package database

import (
	"fmt"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"smart-ledger-server/internal/config"
	"smart-ledger-server/migrations"
)

var db *gorm.DB

// Init 初始化数据库连接
func Init(cfg *config.DatabaseConfig, log *zap.Logger) error {
	var gormLogger logger.Interface
	if cfg.Host == "" {
		return fmt.Errorf("数据库配置不完整")
	}

	// 设置GORM日志级别
	gormLogger = logger.Default.LogMode(logger.Info)

	var err error
	db, err = gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
		Logger:                                   gormLogger,
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取数据库实例失败: %w", err)
	}

	// 设置连接池
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	log.Info("数据库连接成功",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("dbname", cfg.DBName),
	)

	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return db
}

// AutoMigrate 使用 goose 执行数据库迁移
func AutoMigrate() error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取数据库实例失败: %w", err)
	}

	return migrations.Run(sqlDB)
}

// Close 关闭数据库连接
func Close() error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
