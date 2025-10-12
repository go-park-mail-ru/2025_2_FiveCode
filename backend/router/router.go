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

func NewRouter(s *store.Store, conf *config.Config) http.Handler {

	r := mux.NewRouter()

	authR := authRepository.NewAuthRepository(s)
	userR := userRepository.NewUserRepository(s)
	notesR := notesRepository.NewNotesRepository(s)

	authUC := authUsecase.NewAuthUsecase(authR)
	userUC := userUsecase.NewUserUsecase(userR)
	notesUC := notesUsecase.NewNotesUsecase(notesR)

	authD := authDelivery.NewAuthDelivery(authUC, time.Duration(conf.Cookie.SessionDuration)*24*time.Hour)
	userD := userDelivery.NewUserDelivery(userUC)
	notesD := notesDelivery.NewNotesDelivery(notesUC)

	api := r.PathPrefix("/api").Subrouter()

	api.HandleFunc("/login", authD.Login).Methods("POST")
	api.HandleFunc("/register", userD.Register).Methods("POST")
	api.HandleFunc("/logout", authD.Logout).Methods("POST")
	api.HandleFunc("/session", userD.GetProfile).Methods("GET")
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	protected := api.PathPrefix("").Subrouter()
	protected.Use(mw.AuthMiddleware(s))
	protected.Use(mw.UserAccessMiddleware())
	protected.HandleFunc("/user/{user_id}/notes", notesD.GetAllNotes).Methods("GET")

	return mw.CORS(r)
}
