package main

import (
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"log"
	"os"
)

var config = struct {
	port      int
	mediaDir  string
	chunkSize int64
	redisAddr string
}{
	port:      6969,
	mediaDir:  os.Getenv("MEDIA_DIR"),
	chunkSize: 10 << 20, // 10 MB
	redisAddr: "127.0.0.1:6379",
}
var db *gorm.DB = nil

func main() {
	// load configuration
	// TODO
	// set database connection
	database, err := GetDatabase()
	if err != nil || database == nil {
		log.Fatal("Cannot connect database")
	}
	// Migrate the schema
	err = database.AutoMigrate(&Share{})
	if err != nil {
		log.Fatal("Cannot migrate database")
	}
	err = database.AutoMigrate(&Attachment{})
	if err != nil {
		log.Fatal("Cannot migrate database")
	}
	db = database
	if db == nil {
		log.Fatal("Can't connect to database")
	}
	// background job server
	go StartBackgroundWorker()
	// start
	fmt.Println("Let's go!!!")
	//ConfigurePool()
	ConfigureRoutes()
}

func PrettyPrint(i interface{}) {
	b, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(b))
}
