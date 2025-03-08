package pg

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type queryOptions struct{}

type Store struct {
	db      *gorm.DB
	options *queryOptions
	cache   cache.Cache
}

func (s *Store) GetDB() *gorm.DB {
	return s.db
}

func DSN(host, port, user, password, db string) string {
	timezone := os.Getenv("TZ")
	if timezone == "" {
		timezone = "UTC"
	}

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=%s", host, port, user, password, db, timezone)
}

func New(ctx context.Context, dsn string, cache cache.Cache) (store.Store, error) {
	db, err := connect(ctx, dsn)
	if err != nil {
		return nil, err
	}

	store := &Store{db: db, cache: cache, options: &queryOptions{}}

	return store, nil
}

func connect(ctx context.Context, dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(
		postgres.Open(dsn),
		&gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	if err != nil {
		panic(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}

	if err := sqlDB.Ping(); err != nil {
		panic(err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}
