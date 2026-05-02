package rdb

import (
	"context"
	"database/sql"
)

type healthChecker struct {
	name string
	db   *sql.DB
}

func (h *healthChecker) Name() string {
	return h.name
}

func (h *healthChecker) Check(ctx context.Context) error {
	return h.db.PingContext(ctx)
}
