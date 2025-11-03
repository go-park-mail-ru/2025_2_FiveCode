package initialize

import (
	authDelivery "backend/auth/delivery"
	authRepository "backend/auth/repository"
	authUsecase "backend/auth/usecase"
	blocksDelivery "backend/blocks/delivery"
	blocksRepository "backend/blocks/repository"
	blocksUsecase "backend/blocks/usecase"
	"backend/config"
	notesDelivery "backend/notes/delivery"
	notesRepository "backend/notes/repository"
	notesUsecase "backend/notes/usecase"
	"backend/store"
	userDelivery "backend/user/delivery"
	userRepository "backend/user/repository"
	userUsecase "backend/user/usecase"
	"net/http"
	"time"
)

type AuthDeliveryInterface interface {
	Login(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
}

type UserDeliveryInterface interface {
	Register(w http.ResponseWriter, r *http.Request)
	GetProfile(w http.ResponseWriter, r *http.Request)
	GetProfileBySession(w http.ResponseWriter, r *http.Request)
	UpdateProfile(w http.ResponseWriter, r *http.Request)
}

type NotesDeliveryInterface interface {
	GetAllNotes(w http.ResponseWriter, r *http.Request)
	CreateNote(w http.ResponseWriter, r *http.Request)
	GetNoteById(w http.ResponseWriter, r *http.Request)
	UpdateNote(w http.ResponseWriter, r *http.Request)
	DeleteNote(w http.ResponseWriter, r *http.Request)
}

type BlocksDeliveryInterface interface {
	CreateBlock(w http.ResponseWriter, r *http.Request)
	GetBlocks(w http.ResponseWriter, r *http.Request)
	GetBlock(w http.ResponseWriter, r *http.Request)
	UpdateBlock(w http.ResponseWriter, r *http.Request)
	DeleteBlock(w http.ResponseWriter, r *http.Request)
	UpdateBlockPosition(w http.ResponseWriter, r *http.Request)
}

type Deliveries struct {
	AuthDelivery   AuthDeliveryInterface
	UserDelivery   UserDeliveryInterface
	NotesDelivery  NotesDeliveryInterface
	BlocksDelivery BlocksDeliveryInterface
}

func InitDeliveries(s *store.Store, conf *config.Config) *Deliveries {
	layers := &Deliveries{}

	authR := authRepository.NewAuthRepository(s)
	authUC := authUsecase.NewAuthUsecase(authR)
	layers.AuthDelivery = authDelivery.NewAuthDelivery(authUC, time.Duration(conf.Auth.Cookie.SessionDuration)*24*time.Hour)

	userR := userRepository.NewUserRepository(s)
	userUC := userUsecase.NewUserUsecase(userR, authR)
	layers.UserDelivery = userDelivery.NewUserDelivery(userUC)

	notesR := notesRepository.NewNotesRepository(s)
	notesUC := notesUsecase.NewNotesUsecase(notesR)
	layers.NotesDelivery = notesDelivery.NewNotesDelivery(notesUC)

	blocksR := blocksRepository.NewBlocksRepository(s)
	blocksUC := blocksUsecase.NewBlocksUsecase(blocksR, notesR)
	layers.BlocksDelivery = blocksDelivery.NewBlocksDelivery(blocksUC)

	return layers
}
