package orm

import (
	"gorm.io/gorm"
)

type TransactionManager interface {
	Transactional(fn func(tx *gorm.DB) error) error
}

type transactionManager struct {
	db *gorm.DB
}

func NewTransactionManager(db *gorm.DB) (TransactionManager, error) {
	return &transactionManager{db: db}, nil
}

func (t transactionManager) Transactional(fn func(tx *gorm.DB) error) error {
	err := t.db.Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
	if err != nil {
		return err
	}

	return nil
}
