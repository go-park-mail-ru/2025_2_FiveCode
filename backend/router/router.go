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

	notes := api.PathPrefix("").Subrouter()
	notes.Use(mw.AuthMiddleware(s))
	// ============ ЗАМЕТКИ ============
	// Получить список всех заметок (только метаданные)
	notes.HandleFunc("/notes", deliveries.NotesDelivery.GetAllNotes).Methods("GET")
	// Создать новую заметку
	notes.HandleFunc("/notes", deliveries.NotesDelivery.CreateNote).Methods("POST")
	// Получить заметку целиком (метаданные + все блоки)
	notes.HandleFunc("/notes/{note_id}", deliveries.NotesDelivery.GetNoteById).Methods("GET")
	// Обновить метаданные заметки
	notes.HandleFunc("/notes/{note_id}", deliveries.NotesDelivery.UpdateNote).Methods("PUT")
	// Удалить заметку
	notes.HandleFunc("/notes/{note_id}", deliveries.NotesDelivery.DeleteNote).Methods("DELETE")

	blocks := api.PathPrefix("").Subrouter()
	blocks.Use(mw.AuthMiddleware(s))
	// ============ БЛОКИ ============
	// Создать пустой блок (after_block_id в body)
	blocks.HandleFunc("/notes/{note_id}/blocks", deliveries.BlocksDelivery.CreateBlock).Methods("POST")
	// Получить все блоки заметки
	blocks.HandleFunc("/notes/{note_id}/blocks", deliveries.BlocksDelivery.GetBlocks).Methods("GET")
	// Получить один блок
	blocks.HandleFunc("/blocks/{block_id}", deliveries.BlocksDelivery.GetBlock).Methods("GET")
	// Обновить блок (текст + форматы)
	blocks.HandleFunc("/blocks/{block_id}", deliveries.BlocksDelivery.UpdateBlock).Methods("PATCH")
	// Удалить блок
	blocks.HandleFunc("/blocks/{block_id}", deliveries.BlocksDelivery.DeleteBlock).Methods("DELETE")
	// Изменить позицию блока (after_block_id в body)
	blocks.HandleFunc("/blocks/{block_id}/position", deliveries.BlocksDelivery.UpdateBlockPosition).Methods("PUT")

	return mw.CORS(r)
}
