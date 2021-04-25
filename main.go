package main

import (
	"flag"
	"github.com/chiefsend/api/background"
	"github.com/chiefsend/api/controllers"
	m "github.com/chiefsend/api/models"
	"github.com/joho/godotenv"
	"log"
	"os"
	"path/filepath"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	{
		autoMigrate := flag.Bool("auto-migrate", false, "pass -auto-migrate=true if you want gorm to auto migrate the database")
		flag.Parse()
		if *autoMigrate {
			// set/test database connection
			db, err := m.GetDatabase()
			if err != nil || db == nil {
				log.Fatal("Cannot connect database")
			}
			// Migrate the schema (temporary)
			if err := db.AutoMigrate(&m.Share{}); err != nil {
				log.Fatal("Cannot migrate database")
			}
			if err := db.AutoMigrate(&m.Attachment{}); err != nil {
				log.Fatal("Cannot migrate database")
			}
		}
	}
	// check if file structure is there
	if err := os.MkdirAll(filepath.Join(os.Getenv("MEDIA_DIR"), "temp"), os.ModePerm); err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(os.Getenv("MEDIA_DIR"), "data"), os.ModePerm); err != nil {
		log.Fatal(err)
	}
	// start the server(s)
	log.Print("Starting background workers")
	go background.StartBackgroundWorkers()
	log.Print("Starting Server")
	controllers.StartServer()
}
