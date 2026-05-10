package rdb

import (
	"context"
	"database/sql"
	"fmt"
)

type healthChecker struct {
	name string
	db   *sql.DB
}

func (h *healthChecker) Name() string {
	return h.name
}

func (h *healthChecker) Check(ctx context.Context) error {
	if err := h.db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}
	return nil
}
