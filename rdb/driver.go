package rdb

import (
	"fmt"

	"go.uber.org/fx"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PGConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type PGParams struct {
	fx.In
	Cfg PGConfig
}

func NewPGDriver(p PGParams) gorm.Dialector {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.Cfg.Host,
		p.Cfg.Port,
		p.Cfg.User,
		p.Cfg.Password,
		p.Cfg.DBName,
		p.Cfg.SSLMode,
	)

	dialector := postgres.Open(dsn)

	return dialector
}
