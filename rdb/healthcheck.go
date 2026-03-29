package rdb

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/bencoronard/demo-go-common-libs/actuator"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

type dbHealthChecker struct {
	name string
	db   *sql.DB
}

func (d *dbHealthChecker) Name() string {
	return d.name
}

func (d *dbHealthChecker) Check(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

type DBHealthCheckerParams struct {
	fx.In
	DB *gorm.DB
}

func NewDBHealthChecker(p DBHealthCheckerParams) (actuator.HealthChecker, error) {
	sqlDB, err := p.DB.DB()
	if err != nil {
		return nil, err
	}

	name := fmt.Sprintf("rdb_%s", p.DB.Dialector.Name())

	return &dbHealthChecker{
		name: name,
		db:   sqlDB,
	}, nil
}
