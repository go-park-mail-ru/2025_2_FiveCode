package router

import (
	"backend/handler"
	mw "backend/middleware"
	"backend/store"
	"net/http"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "backend/docs"
)

func NewRouter(s *store.Store) http.Handler {
	h := handler.NewHandler(s)

	r := mux.NewRouter()

	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/register", h.Register).Methods("POST")
	api.HandleFunc("/login", h.Login).Methods("POST")
	api.HandleFunc("/logout", h.Logout).Methods("POST")
	api.HandleFunc("/session", h.CheckSession).Methods("GET")
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	protected := api.PathPrefix("").Subrouter()
	protected.Use(mw.MakeAuthMiddleware(s))
	protected.Use(mw.ValidateUserAccess(s))
	protected.HandleFunc("/user/{user_id}/notes", h.ListNotes).Methods("GET")

	return mw.CORS(r)
}
