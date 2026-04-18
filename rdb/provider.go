package rdb

import (
	"context"
	"fmt"
	"time"

	"github.com/bencoronard/demo-go-common-libs/actuator"
	"go.uber.org/fx"

	"gorm.io/gorm"
)

type DbConfig struct {
	MaxOpenConns int
	MaxIdleConns int
	ConnMaxTTL   time.Duration
	IdleTimeout  time.Duration
}

type dbParams struct {
	fx.In
	Lc  fx.Lifecycle
	Dl  gorm.Dialector
	Cfg DbConfig
}

func NewDb(p dbParams) (*gorm.DB, error) {
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

	sqlDB.SetMaxOpenConns(p.Cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(p.Cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(p.Cfg.ConnMaxTTL)
	sqlDB.SetConnMaxIdleTime(p.Cfg.IdleTimeout)

	p.Lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return sqlDB.PingContext(ctx)
		},
		OnStop: func(_ context.Context) error {
			return sqlDB.Close()
		},
	})

	return db, nil
}

type TransactionManager interface {
	Transactional(ctx context.Context, fn func(tx *gorm.DB) error) error
}

type tmParams struct {
	fx.In
	DB *gorm.DB
}

func NewTransactionManager(p tmParams) (TransactionManager, error) {
	return &transactionManager{db: p.DB}, nil
}

type healthCheckerParams struct {
	fx.In
	DB *gorm.DB
}

func NewDbHealthChecker(p healthCheckerParams) (actuator.HealthChecker, error) {
	sqlDB, err := p.DB.DB()
	if err != nil {
		return nil, err
	}

	name := fmt.Sprintf("rdb_%s", p.DB.Dialector.Name())

	return &healthChecker{
		name: name,
		db:   sqlDB,
	}, nil
}
