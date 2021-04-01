package models

import (
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"os"
	"strings"
)

func GetDatabase() (*gorm.DB, error) {
	dsn := os.Getenv("DATABASE_URI")
	split := strings.SplitN(dsn, ":", 2)
	if split[0] == "postgres" {
		return gorm.Open(postgres.Open(dsn), &gorm.Config{})
	} else if split[0] == "sqlserver" {
		return gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
	} else if split[0] == "sqlite" || split[0] == "file" {
		return gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	}
	return nil, nil
}
