package rdb

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/bencoronard/demo-go-common-libs/actuator"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

type healthChecker struct {
	name string
	db   *sql.DB
}

func (d *healthChecker) Name() string {
	return d.name
}

func (d *healthChecker) Check(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

type HealthCheckerParams struct {
	fx.In
	DB *gorm.DB
}

func NewDBHealthChecker(p HealthCheckerParams) (actuator.HealthChecker, error) {
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
