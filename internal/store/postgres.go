package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Store struct {
	GORM *gorm.DB
	Pool *pgxpool.Pool
}

func Open(ctx context.Context, dsn string) (*Store, error) {
	gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("ouverture gorm: %w", err)
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("ouverture pgxpool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping base: %w", err)
	}

	return &Store{GORM: gdb, Pool: pool}, nil
}

func (s *Store) Close() error {
	if s.Pool != nil {
		s.Pool.Close()
	}
	if s.GORM != nil {
		db, err := s.GORM.DB()
		if err != nil {
			return err
		}
		return db.Close()
	}
	return nil
}
