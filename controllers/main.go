package controllers

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"log"
	"net/http"
	"os"
)

func ConfigureRoutes() {
	router := mux.NewRouter()
	handler := cors.Default().Handler(router)

	router.Handle("/shares", EndpointREST(AllShares)).Methods("GET")
	router.Handle("/shares", EndpointREST(OpenShare)).Methods("POST")

	router.Handle("/share/{id}", EndpointREST(GetShare)).Methods("GET")
	router.Handle("/share/{id}", EndpointREST(CloseShare)).Methods("POST")

	router.Handle("/share/{id}/attachments", EndpointREST(UploadAttachment)).Methods("POST")

	router.Handle("/share/{id}/attachment/{att}", EndpointREST(DownloadFile)).Methods("GET")
	router.Handle("/share/{id}/zip", EndpointREST(DownloadZip)).Methods("GET")

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), handler))
}