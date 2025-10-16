package router

import (
	authDelivery "backend/auth/delivery"
	authRepository "backend/auth/repository"
	authUsecase "backend/auth/usecase"
	"backend/config"
	mw "backend/middleware"
	notesDelivery "backend/notes/delivery"
	notesRepository "backend/notes/repository"
	notesUsecase "backend/notes/usecase"
	"backend/store"
	userDelivery "backend/user/delivery"
	userRepository "backend/user/repository"
	userUsecase "backend/user/usecase"
	"net/http"
	"time"

	_ "backend/docs"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

func NewRouter(s *store.Store, deliveries *Deliveries) http.Handler {
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

	return mw.CORS(r)
}

type Deliveries struct {
	AuthDelivery  *authDelivery.AuthDelivery
	UserDelivery  *userDelivery.UserDelivery
	NotesDelivery *notesDelivery.NotesDelivery
}

func InitDeliveries(s *store.Store, conf *config.Config) *Deliveries {
	layers := &Deliveries{}

	authR := authRepository.NewAuthRepository(s)
	authUC := authUsecase.NewAuthUsecase(authR)
	layers.AuthDelivery = authDelivery.NewAuthDelivery(authUC, time.Duration(conf.Cookie.SessionDuration)*24*time.Hour)

	userR := userRepository.NewUserRepository(s)
	userUC := userUsecase.NewUserUsecase(userR)
	layers.UserDelivery = userDelivery.NewUserDelivery(userUC)

	notesR := notesRepository.NewNotesRepository(s)
	notesUC := notesUsecase.NewNotesUsecase(notesR)
	layers.NotesDelivery = notesDelivery.NewNotesDelivery(notesUC)

	return layers
}
