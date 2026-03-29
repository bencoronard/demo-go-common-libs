package rdb

import (
	"context"
	"time"

	"go.uber.org/fx"

	"gorm.io/gorm"
)

type DBConfig struct {
	CPCap         int
	CPIdleMin     int
	CPConnTTL     time.Duration
	CPIdleTimeout time.Duration
}

type DBParams struct {
	fx.In
	Lc  fx.Lifecycle
	Dl  gorm.Dialector
	Cfg *DBConfig
}

func NewDB(p DBParams) (*gorm.DB, error) {
	db, err := gorm.Open(p.Dl, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(p.Cfg.CPCap)
	sqlDB.SetMaxIdleConns(p.Cfg.CPIdleMin)
	sqlDB.SetConnMaxLifetime(p.Cfg.CPConnTTL)
	sqlDB.SetConnMaxIdleTime(p.Cfg.CPIdleTimeout)

	p.Lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return sqlDB.PingContext(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return sqlDB.Close()
		},
	})

	return db, nil
}
