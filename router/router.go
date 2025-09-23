package router

import (
	"backend/handler"
	mw "backend/middleware"
	"backend/store"
	"fmt"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

const frontendHost = "localhost"
const frontendPort = "3000"

func NewRouter(s *store.Store) http.Handler {
	h := handler.NewHandler(s)

	r := mux.NewRouter()

	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/register", h.Register).Methods("POST")
	api.HandleFunc("/login", h.Login).Methods("POST")
	api.HandleFunc("/logout", h.Logout).Methods("POST")

	protected := api.PathPrefix("").Subrouter()
	protected.Use(mw.MakeAuthMiddleware(s))
	protected.HandleFunc("/user/{id}/notes", h.ListNotes).Methods("GET")

	corsOpts := handlers.AllowedOrigins([]string{fmt.Sprintf("http://%s:%s", frontendHost, frontendPort)})
	corsMethods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	corsHeaders := handlers.AllowedHeaders([]string{"Content-Type", "Authorization"})

	return handlers.CORS(corsOpts, corsMethods, corsHeaders)(r)
}
