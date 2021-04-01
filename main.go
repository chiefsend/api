package main

import (
	"encoding/json"
	"fmt"
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

func main() {
	// load configuration
	// TODO
	// test database connection
	db, err := GetDatabase()
	if err != nil {
		log.Fatal("Cannot connect database")
	}
	// Migrate the schema
	err = db.AutoMigrate(&Share{})
	if err != nil {
		log.Fatal("Cannot migrate database")
	}
	err = db.AutoMigrate(&Attachment{})
	if err != nil {
		log.Fatal("Cannot migrate database")
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
