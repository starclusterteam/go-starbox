package db

import (
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var ErrUniqueConstraintViolation = errors.New("unique constraint violation")

func TranslateError(err error, db *gorm.DB) error {
	if err == nil {
		return nil
	}

	switch db.Dialector.Name() {
	case "sqlite":
		if strings.HasPrefix(err.Error(), "UNIQUE constraint failed:") {
			return ErrUniqueConstraintViolation
		}
		return err
	case "mysql":
		if err, ok := err.(*mysql.MySQLError); ok {
			if err.Number == 1062 {
				return ErrUniqueConstraintViolation
			}
		}
		return err
	default:
		return err
	}
}
