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
}{
	port:      6969,
	mediaDir:  os.Getenv("MEDIA_DIR"),
	chunkSize: 10 << 20, // 10 MB
}

func main() {
	// test database connection
	db, err := GetDatabase()
	if err != nil {
		log.Fatal("Cannot connect database")
	}
	// Migrate the schema
	db.AutoMigrate(&Share{})
	db.AutoMigrate(&Attachment{})
	// start
	fmt.Println("Let's go!!!")
	ConfigureRoutes()
	//ConfigurePool()
}

func PrettyPrint(i interface{}) {
	b, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(b))
}
