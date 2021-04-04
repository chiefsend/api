package main

import (
	"fmt"
	"github.com/chiefsend/api/background"
	"github.com/chiefsend/api/controllers"
	g "github.com/chiefsend/api/globals"
	m "github.com/chiefsend/api/models"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"log"
	"net/http"
)

func main() {
	// load configuration
	g.LoadConfig()
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
	// start the server(s)
	fmt.Println("Lets go!")
	go background.StartBackgroundWorkers()
	ConfigureRoutes()
}

func ConfigureRoutes() {
	router := mux.NewRouter()
	handler := cors.Default().Handler(router)

	router.Handle("/shares", controllers.EndpointREST(controllers.AllShares)).Methods("GET")
	router.Handle("/shares", controllers.EndpointREST(controllers.OpenShare)).Methods("POST")

	router.Handle("/share/{id}", controllers.EndpointREST(controllers.GetShare)).Methods("GET")
	router.Handle("/share/{id}", controllers.EndpointREST(controllers.CloseShare)).Methods("POST")

	router.Handle("/share/{id}/attachments", controllers.EndpointREST(controllers.UploadAttachment)).Methods("POST")

	router.Handle("/share/{id}/attachment/{att}", controllers.EndpointREST(controllers.DownloadFile)).Methods("GET")
	router.Handle("/share/{id}/zip", controllers.EndpointREST(controllers.DownloadZip)).Methods("GET")

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", g.Conf.Port), handler))
}
