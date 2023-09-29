package db

import (
	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/starclusterteam/go-starbox/config"
)

func NewDatabaseConnection(dsn string) (*gorm.DB, error) {
	logLevelStr := config.String("MYSQL_LOG_LEVEL", "silent")
	var logLevel logger.LogLevel
	switch logLevelStr {
	case "silent":
		logLevel = logger.Silent
	case "error":
		logLevel = logger.Error
	case "warn":
		logLevel = logger.Warn
	case "info":
		logLevel = logger.Info
	}

	d, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to database")
	}

	return d, nil
}
