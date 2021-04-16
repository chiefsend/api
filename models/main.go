package models

import (
	"errors"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"os"
	"strings"
)

var db *gorm.DB = nil

// returns the database connection. Creates one if doesn't exist.
func GetDatabase() (*gorm.DB, error) {
	if db != nil {
		return db, nil
	}

	dsn := os.Getenv("DATABASE_URI")
	split := strings.SplitN(dsn, ":", 2)
	if split[0] == "postgres" {  // TODO das ist keine PostgreSQL DSN siehe: https://gorm.io/docs/connecting_to_the_database.html#PostgreSQL
		return gorm.Open(postgres.Open(dsn), &gorm.Config{})
	} else if split[0] == "sqlserver" {
		return gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
	} else if split[0] == "sqlite" || split[0] == "file" {
		return gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	}else {
		return nil, errors.New("invalid DSN")
	}
}
