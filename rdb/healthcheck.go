package rdb

import (
	"context"
	"database/sql"
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
