package db

import (
	"context"

	"github.com/pkg/errors"
	"github.com/starclusterteam/go-starbox/log"
	"gorm.io/gorm"
)

type BaseRepository struct {
	db     *gorm.DB
	Tables []TableI
}

func NewBaseRepository(db *gorm.DB, tables []TableI) *BaseRepository {
	return &BaseRepository{
		db:     db,
		Tables: tables,
	}
}

func (r *BaseRepository) Init(ctx context.Context) error {
	for _, t := range r.Tables {
		log.Infof("Migrating mysql table %s", t.TableName())

		if err := r.db.WithContext(ctx).AutoMigrate(t); err != nil {
			return errors.Wrapf(err, "failed to auto migrate %s", t.TableName())
		}

		log.Infof("Migrated mysql table %s", t.TableName())
	}

	return nil
}

// nolint:revive
func (r *BaseRepository) CreateIndexes(ctx context.Context) error {
	return nil
}
