package rdb

import (
	"context"

	"gorm.io/gorm"
)

type transactionManager struct {
	db *gorm.DB
}

func (t transactionManager) Transactional(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}
