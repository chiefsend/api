package models

import (
	"errors"
	"gorm.io/driver/clickhouse"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"os"
)

var db *gorm.DB = nil

// returns the database connection. Opens one if doesn't exist.
func GetDatabase() (*gorm.DB, error) {
	if db != nil {
		return db, nil
	}

	dsn := os.Getenv("DATABASE_URI")
	switch os.Getenv("DATABASE_DIALECT") {
	case "mysql":
		return gorm.Open(mysql.Open(dsn), &gorm.Config{})
	case "postgres":
		return gorm.Open(postgres.Open(dsn), &gorm.Config{})
	case "sqlite":
		return gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	case "mssql":
		return gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
	case "clickhouse":
		return gorm.Open(clickhouse.Open(dsn), &gorm.Config{})
	default:
		return nil, errors.New("invalid DSN")
	}
}
