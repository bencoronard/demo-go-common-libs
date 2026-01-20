package orm

import (
	"context"

	"gorm.io/gorm"
)

type TransactionManager interface {
	Transactional(ctx context.Context, fn func(tx *gorm.DB) error) error
}

type transactionManager struct {
	db *gorm.DB
}

func NewTransactionManager(db *gorm.DB) (TransactionManager, error) {
	return &transactionManager{db: db}, nil
}

func (t transactionManager) Transactional(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}
