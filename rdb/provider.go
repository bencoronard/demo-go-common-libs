package rdb

import (
	"context"
	"fmt"
	"time"

	"github.com/bencoronard/demo-go-common-libs/actuator"
	"go.uber.org/fx"

	"gorm.io/gorm"
)

type DBConfig struct {
	MaxOpenConns int
	MaxIdleConns int
	ConnTTL      time.Duration
	IdleTimeout  time.Duration
}

type dbParams struct {
	fx.In
	Lifecycle fx.Lifecycle
	Dialector gorm.Dialector
	Config    DBConfig
}

func NewDB(p dbParams) (*gorm.DB, error) {
	db, err := gorm.Open(p.Dialector, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(p.Config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(p.Config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(p.Config.ConnTTL)
	sqlDB.SetConnMaxIdleTime(p.Config.IdleTimeout)

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return sqlDB.PingContext(ctx)
		},
		OnStop: func(_ context.Context) error {
			return sqlDB.Close()
		},
	})

	return db, nil
}

type healthCheckerParams struct {
	fx.In
	DB *gorm.DB
}

func NewHealthChecker(p healthCheckerParams) (actuator.HealthChecker, error) {
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
