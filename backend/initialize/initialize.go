package initialize

import (
	authDelivery "backend/auth/delivery"
	authRepository "backend/auth/repository"
	authUsecase "backend/auth/usecase"
	"backend/config"
	notesDelivery "backend/notes/delivery"
	notesRepository "backend/notes/repository"
	notesUsecase "backend/notes/usecase"
	profileDelivery "backend/profile/delivery"
	profileRepository "backend/profile/repository"
	profileUsecase "backend/profile/usecase"
	"backend/store"
	userDelivery "backend/user/delivery"
	userRepository "backend/user/repository"
	userUsecase "backend/user/usecase"
	"time"
)

type Deliveries struct {
	AuthDelivery    *authDelivery.AuthDelivery
	UserDelivery    *userDelivery.UserDelivery
	NotesDelivery   *notesDelivery.NotesDelivery
	ProfileDelivery *profileDelivery.ProfileDelivery
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

	profileR := profileRepository.NewProfileRepository(s)
	profileUC := profileUsecase.NewProfileUsecase(profileR)
	layers.ProfileDelivery = profileDelivery.NewProfileDelivery(profileUC)

	return layers
}
