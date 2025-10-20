package router

import (
	"backend/initialize"
	mw "backend/middleware"
	"backend/store"
	"net/http"

	_ "backend/docs"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

func NewRouter(s *store.Store, deliveries *initialize.Deliveries) http.Handler {
	r := mux.NewRouter()

	api := r.PathPrefix("/api").Subrouter()

	api.HandleFunc("/login", deliveries.AuthDelivery.Login).Methods("POST")
	api.HandleFunc("/register", deliveries.UserDelivery.Register).Methods("POST")
	api.HandleFunc("/logout", deliveries.AuthDelivery.Logout).Methods("POST")
	api.HandleFunc("/session", deliveries.UserDelivery.GetProfile).Methods("GET")
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	protected := api.PathPrefix("").Subrouter()
	protected.Use(mw.AuthMiddleware(s))
	protected.Use(mw.UserAccessMiddleware())
	protected.HandleFunc("/user/{user_id}/notes", deliveries.NotesDelivery.GetAllNotes).Methods("GET")

	profile := api.PathPrefix("").Subrouter()
	profile.Use(mw.AuthMiddleware(s))
	profile.HandleFunc("/profile", deliveries.ProfileDelivery.GetProfile).Methods("GET")
	profile.HandleFunc("/profile", deliveries.ProfileDelivery.UpdateProfile).Methods("PUT")

	return mw.CORS(r)
}
