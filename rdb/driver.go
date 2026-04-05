package rdb

import (
	"fmt"

	"go.uber.org/fx"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DriverConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	UseSSL   bool
}

type DriverParams struct {
	fx.In
	Cfg DriverConfig
}

func NewPGDriver(p DriverParams) gorm.Dialector {
	sslMode := "disable"
	if p.Cfg.UseSSL {
		sslMode = "require"
	}

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.Cfg.Host,
		p.Cfg.Port,
		p.Cfg.User,
		p.Cfg.Password,
		p.Cfg.DBName,
		sslMode,
	)

	dialector := postgres.Open(dsn)

	return dialector
}
