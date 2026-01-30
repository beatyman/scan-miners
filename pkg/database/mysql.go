package database

import (
	"github.com/beatyman/scan-miners/config"
	"github.com/beatyman/scan-miners/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewMySQLConnection(cfg *config.Config) (*gorm.DB, error) {
	dsn := cfg.MySQL.DSN()
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Log.Error("Failed to connect to database", zap.Error(err))
		return nil, err
	}

	logger.Log.Info("Successfully connected to database")
	return db, nil
}
