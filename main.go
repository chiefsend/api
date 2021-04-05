package main

import (
	"flag"
	"fmt"
	"github.com/chiefsend/api/controllers"
	m "github.com/chiefsend/api/models"
	"github.com/joho/godotenv"
	"log"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	autoMigrate := flag.Bool("auto-migrate", false, "pass -auto-migrate=true if you want gorm to auto migrate the database")
	flag.Parse()
	if *autoMigrate {
		// set/test database connection
		db, err := m.GetDatabase()
		if err != nil || db == nil {
			log.Fatal("Cannot connect database")
		}
		// Migrate the schema (temporary) FIXME: nur mit passender flag?
		if err := db.AutoMigrate(&m.Share{}); err != nil {
			log.Fatal("Cannot migrate database")
		}
		if err := db.AutoMigrate(&m.Attachment{}); err != nil {
			log.Fatal("Cannot migrate database")
		}
	}
	// start the server(s)
	fmt.Println("Lets go!")
	//go background.StartBackgroundWorkers()
	controllers.ConfigureRoutes()
}

