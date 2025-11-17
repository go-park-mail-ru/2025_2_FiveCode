package initialize

import (
	"backend/config"
	"backend/store"
	authPB "backend/gen/go/auth"
	"backend/pkg/gateway/delivery"
	"backend/pkg/auth/repository"
	userDelivery "backend/pkg/gateway/delivery"
	userRepository "backend/pkg/user/repository"
	userUsecase "backend/pkg/user/usecase"
	notesDelivery "backend/pkg/gateway/notes/delivery"
	notesRepository "backend/pkg/gateway/notes/repository"
	notesUsecase "backend/pkg/gateway/notes/usecase"
	blocksDelivery "backend/pkg/gateway/blocks/delivery"
	blocksRepository "backend/pkg/gateway/blocks/repository"
	blocksUsecase "backend/pkg/gateway/blocks/usecase"
	fileDelivery "backend/pkg/gateway/file/delivery"
	fileRepository "backend/pkg/gateway/file/repository"
	fileUsecase "backend/pkg/gateway/file/usecase"
	"fmt"
	"github.com/rs/zerolog/log"
	"net/http"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthDeliveryInterface interface {
	Login(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
	Register(w http.ResponseWriter, r *http.Request)
	// GetCSRFToken(w http.ResponseWriter, r *http.Request)
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

	AuthClient authPB.AuthClient
}

func InitDeliveries(s *store.Store, conf *config.Config) *Deliveries {
	authServiceAddr := fmt.Sprintf("localhost:%d", conf.Services["auth"].GrpcPort)
	authServiceConn, err := grpc.Dial(
		authServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to auth service")
	}
	// defer authServiceConn.Close()

	authGRPCClient := authPB.NewAuthClient(authServiceConn)

	layers := &Deliveries{
		AuthClient: authGRPCClient,
	}
	layers.AuthDelivery = delivery.NewAuthDelivery(authGRPCClient, time.Duration(conf.Auth.Cookie.SessionDuration)*24*time.Hour)

	authR_legacy := repository.NewAuthRepository(s.Postgres.DB, s.Redis.Client)
	userR := userRepository.NewUserRepository(s.Postgres.DB)
	userUC := userUsecase.NewUserUsecase(userR, authR_legacy)
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

	return layers
}
