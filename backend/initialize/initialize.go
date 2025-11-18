package initialize

import (
	"backend/config"
	authPB "backend/gen/go/auth"
	userPB "backend/gen/go/user"
	blocksDelivery "backend/pkg/gateway/blocks/delivery"
	blocksRepository "backend/pkg/gateway/blocks/repository"
	blocksUsecase "backend/pkg/gateway/blocks/usecase"
	authDelivery "backend/pkg/gateway/delivery/auth"
	userDelivery "backend/pkg/gateway/delivery/user"
	fileDelivery "backend/pkg/gateway/file/delivery"
	fileRepository "backend/pkg/gateway/file/repository"
	fileUsecase "backend/pkg/gateway/file/usecase"
	notesDelivery "backend/pkg/gateway/notes/delivery"
	notesRepository "backend/pkg/gateway/notes/repository"
	notesUsecase "backend/pkg/gateway/notes/usecase"
	"backend/store"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthDeliveryInterface interface {
	Login(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
	Register(w http.ResponseWriter, r *http.Request)
	GetCSRFToken(w http.ResponseWriter, r *http.Request)
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
	UserClient userPB.UserServiceClient
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

	authGRPCClient := authPB.NewAuthClient(authServiceConn)

	userServiceAddr := fmt.Sprintf("localhost:%d", conf.Services["users"].GrpcPort)
	userServiceConn, err := grpc.Dial(
		userServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to user service")
	}
	userGRPCClient := userPB.NewUserServiceClient(userServiceConn)

	layers := &Deliveries{
		AuthClient: authGRPCClient,
		UserClient: userGRPCClient,
	}
	layers.AuthDelivery = authDelivery.NewAuthDelivery(authGRPCClient, userGRPCClient, time.Duration(conf.Auth.Cookie.SessionDuration)*24*time.Hour)

	layers.UserDelivery = userDelivery.NewUserDelivery(userGRPCClient, authGRPCClient)

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
