package actuator

import (
	"context"
	"database/sql"
	"sync/atomic"
	"time"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

type ActuatorParams struct {
	fx.In
	Lifecycle fx.Lifecycle
	DB        *gorm.DB `optional:"true"`
}

type Actuator interface {
	Liveness() bool
	Readiness() bool
}

type actuatorImpl struct {
	ready atomic.Bool
	db    *sql.DB
}

func NewActuator(p ActuatorParams) (Actuator, error) {
	var err error
	var sqlDB *sql.DB
	if p.DB != nil {
		sqlDB, err = p.DB.DB()
		if err != nil {
			return nil, err
		}
	}

	a := &actuatorImpl{
		db: sqlDB,
	}

	ctx, cancel := context.WithCancel(context.Background())

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go a.monitor(ctx)
			return nil
		},
		OnStop: func(_ context.Context) error {
			cancel()
			return nil
		},
	})

	return a, nil
}

func (a *actuatorImpl) Liveness() bool {
	return true
}

func (a *actuatorImpl) Readiness() bool {
	return a.ready.Load()
}

func (a *actuatorImpl) monitor(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			err := a.check(pingCtx)
			cancel()
			a.ready.Store(err == nil)
		case <-ctx.Done():
			return
		}
	}
}

func (a *actuatorImpl) check(ctx context.Context) error {
	if a.db != nil {
		return a.db.PingContext(ctx)
	}
	return nil
}
