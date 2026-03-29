package rdb

import (
	"context"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

type TransactionManager interface {
	Transactional(ctx context.Context, fn func(tx *gorm.DB) error) error
}

type transactionManager struct {
	db *gorm.DB
}

func (t transactionManager) Transactional(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

type TMParams struct {
	fx.In
	DB *gorm.DB
}

func NewTransactionManager(p TMParams) (TransactionManager, error) {
	return &transactionManager{db: p.DB}, nil
}
