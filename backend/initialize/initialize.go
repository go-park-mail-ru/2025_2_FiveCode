package initialize

import (
	authDelivery "backend/auth/delivery"
	authRepository "backend/auth/repository"
	authUsecase "backend/auth/usecase"
	blocksDelivery "backend/blocks/delivery"
	blocksRepository "backend/blocks/repository"
	blocksUsecase "backend/blocks/usecase"
	"backend/config"
	fileDelivery "backend/file/delivery"
	fileRepository "backend/file/repository"
	fileUsecase "backend/file/usecase"
	notesDelivery "backend/notes/delivery"
	notesRepository "backend/notes/repository"
	notesUsecase "backend/notes/usecase"
	"backend/store"
	ticketDelivery "backend/ticket/delivery"
	ticketRepository "backend/ticket/repository"
	ticketUsecase "backend/ticket/usecase"
	userDelivery "backend/user/delivery"
	userRepository "backend/user/repository"
	userUsecase "backend/user/usecase"
	"net/http"
	"time"
)

type TicketDeliveryInterface interface {
	GetAllTicketsByUserId(w http.ResponseWriter, r *http.Request)
	UpdateTicket(w http.ResponseWriter, r *http.Request)
	GetTicketById(w http.ResponseWriter, r *http.Request)
	CreateTicket(w http.ResponseWriter, r *http.Request)
	GetStatistics(w http.ResponseWriter, r *http.Request)
	GetAllTickets(w http.ResponseWriter, r *http.Request)
	UpdateTicketStatus(w http.ResponseWriter, r *http.Request)
	CreateTicketMessage(w http.ResponseWriter, r *http.Request)
	GetTicketMessages(w http.ResponseWriter, r *http.Request)
}

type AuthDeliveryInterface interface {
	Login(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
	Register(w http.ResponseWriter, r *http.Request)
}

type UserDeliveryInterface interface {
	GetProfile(w http.ResponseWriter, r *http.Request)
	GetProfileBySession(w http.ResponseWriter, r *http.Request)
	UpdateProfile(w http.ResponseWriter, r *http.Request)
	DeleteProfile(w http.ResponseWriter, r *http.Request)
}

type NotesDeliveryInterface interface {
	GetAllNotes(w http.ResponseWriter, r *http.Request)
	CreateNote(w http.ResponseWriter, r *http.Request)
	GetNoteById(w http.ResponseWriter, r *http.Request)
	UpdateNote(w http.ResponseWriter, r *http.Request)
	DeleteNote(w http.ResponseWriter, r *http.Request)
	AddFavorite(w http.ResponseWriter, r *http.Request)
	RemoveFavorite(w http.ResponseWriter, r *http.Request)
}

type BlocksDeliveryInterface interface {
	CreateBlock(w http.ResponseWriter, r *http.Request)
	GetBlocks(w http.ResponseWriter, r *http.Request)
	GetBlock(w http.ResponseWriter, r *http.Request)
	UpdateBlock(w http.ResponseWriter, r *http.Request)
	DeleteBlock(w http.ResponseWriter, r *http.Request)
	UpdateBlockPosition(w http.ResponseWriter, r *http.Request)
}

type FileDeliveryInterface interface {
	UploadFile(w http.ResponseWriter, r *http.Request)
	GetFile(w http.ResponseWriter, r *http.Request)
	DeleteFile(w http.ResponseWriter, r *http.Request)
}

type Deliveries struct {
	AuthDelivery   AuthDeliveryInterface
	UserDelivery   UserDeliveryInterface
	NotesDelivery  NotesDeliveryInterface
	BlocksDelivery BlocksDeliveryInterface
	FileDelivery   FileDeliveryInterface
	TicketDelivery TicketDeliveryInterface
}

func InitDeliveries(s *store.Store, conf *config.Config) *Deliveries {
	layers := &Deliveries{}

	authR := authRepository.NewAuthRepository(s.Postgres.DB, s.Redis.Client)
	authUC := authUsecase.NewAuthUsecase(authR)
	layers.AuthDelivery = authDelivery.NewAuthDelivery(authUC, time.Duration(conf.Auth.Cookie.SessionDuration)*24*time.Hour)

	userR := userRepository.NewUserRepository(s.Postgres.DB)
	userUC := userUsecase.NewUserUsecase(userR, authR)
	layers.UserDelivery = userDelivery.NewUserDelivery(userUC)

	notesR := notesRepository.NewNotesRepository(s.Postgres.DB)
	notesUC := notesUsecase.NewNotesUsecase(notesR)
	layers.NotesDelivery = notesDelivery.NewNotesDelivery(notesUC)

	blocksR := blocksRepository.NewBlocksRepository(s.Postgres.DB)
	blocksUC := blocksUsecase.NewBlocksUsecase(blocksR, notesR)
	layers.BlocksDelivery = blocksDelivery.NewBlocksDelivery(blocksUC)

	fileRepo := fileRepository.NewFileRepository(s.Postgres.DB, s.Minio.Client)
	fileUC := fileUsecase.NewFileUsecase(fileRepo)
	layers.FileDelivery = fileDelivery.NewFileDelivery(fileUC)

	ticketRepo := ticketRepository.NewTicketRepository(s.Postgres.DB)
	ticketUC := ticketUsecase.NewTicketUsecase(ticketRepo)
	layers.TicketDelivery = ticketDelivery.NewTicketDelivery(ticketUC)

	return layers
}
