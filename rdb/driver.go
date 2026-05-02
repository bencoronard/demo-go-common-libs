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

type driverParams struct {
	fx.In
	Config DriverConfig
}

func NewPgDriver(p driverParams) gorm.Dialector {
	sslMode := "disable"
	if p.Config.UseSSL {
		sslMode = "require"
	}

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.Config.Host,
		p.Config.Port,
		p.Config.User,
		p.Config.Password,
		p.Config.DBName,
		sslMode,
	)

	dialector := postgres.Open(dsn)

	return dialector
}
