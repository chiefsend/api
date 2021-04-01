package main

import (
	g "chiefsend-api/globals"
	m "chiefsend-api/models"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"log"
	"net/http"
	"os"
)

func main() {
	// setup logging
	file, err := os.OpenFile("log.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	log.SetOutput(file)
	// load configuration
	g.LoadConfig()
	// set database connection
	database, err := m.GetDatabase()
	if err != nil || database == nil {
		log.Fatal("Cannot connect database")
	}
	g.Db = database
	if g.Db == nil {
		log.Fatal("Can't connect to database")
	}
	// Migrate the schema
	err = g.Db.AutoMigrate(&m.Share{})
	if err != nil {
		log.Fatal("Cannot migrate database")
	}
	err = g.Db.AutoMigrate(&m.Attachment{})
	if err != nil {
		log.Fatal("Cannot migrate database")
	}

	fmt.Println("Lets go!")
	// background job server
	go StartBackgroundWorker()
	// setup routes
	ConfigureRoutes()
}

func ConfigureRoutes() {
	router := mux.NewRouter().StrictSlash(true)
	handler := cors.Default().Handler(router)

	router.Handle("/shares", endpointREST(AllShares)).Methods("GET")
	router.Handle("/shares", endpointREST(OpenShare)).Methods("POST")

	router.Handle("/share/{id}", endpointREST(GetShare)).Methods("GET")
	router.Handle("/share/{id}", endpointREST(CloseShare)).Methods("POST")

	router.Handle("/share/{id}/attachments", endpointREST(UploadAttachment)).Methods("POST")

	router.Handle("/share/{id}/attachment/{att}", endpointREST(DownloadFile)).Methods("GET")
	router.Handle("/share/{id}/zip", endpointREST(DownloadZip)).Methods("GET")

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", g.Conf.Port), handler))
}

func PrettyPrint(i interface{}) {
	b, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		log.Println(err)
	}
	fmt.Println(string(b))
}
