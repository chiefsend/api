package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/chiefsend/api/models"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"log"
	"net/http"
	"os"
)

// ConfigureRoutes sets up the mux router
func configureRoutes(router *mux.Router) {
	router.Handle("/shares", EndpointREST(AllShares)).Methods("GET")
	router.Handle("/shares", EndpointREST(OpenShare)).Methods("POST")

	router.Handle("/share/{id}", EndpointREST(GetShare)).Methods("GET")
	router.Handle("/share/{id}", EndpointREST(CloseShare)).Methods("POST")
	router.Handle("/share/{id}", EndpointREST(DeleteShare)).Methods("DELETE")
	router.Handle("/share/{id}", EndpointREST(UpdateShare)).Methods("PUT")

	router.Handle("/share/{id}/attachments", EndpointREST(UploadAttachment)).Methods("POST")

	router.Handle("/share/{id}/attachment/{att}", EndpointREST(DownloadFile)).Methods("GET")
	router.Handle("/share/{id}/attachment/{att}", EndpointREST(DeleteAttachment)).Methods("DELETE")

	router.Handle("/share/{id}/zip", EndpointREST(DownloadZip)).Methods("GET")

	router.Handle("/shares/stats", EndpointREST(Stats)).Methods("GET")
	router.Handle("/share/{id}/stats", EndpointREST(ShareStats)).Methods("GET")
	router.Handle("/jobs/", EndpointREST(Jobs)).Methods("GET")
}

func StartServer() {
	router := mux.NewRouter()
	handler := cors.AllowAll().Handler(router)
	configureRoutes(router)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), handlers.LoggingHandler(os.Stdout, handler)))
}

func sendJSON(w http.ResponseWriter, res interface{}) *HTTPError {
	// Secure outgoing data
	switch v := res.(type) {
	case models.Share:
		v.Secure()
		res = v
	case []models.Share:
		for i := range v {
			v[i].Secure()
		}
		res = v
	}
	// Send it
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(res)
	if err != nil {
		return &HTTPError{err, "Can't encode data", 500}
	}
	return nil
}
