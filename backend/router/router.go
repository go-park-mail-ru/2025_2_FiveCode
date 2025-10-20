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

	// заметки
	protected.HandleFunc("/notes", deliveries.NotesDelivery.GetAllNotes).Methods("GET")
	protected.HandleFunc("/notes", deliveries.NotesDelivery.CreateNote).Methods("POST")
	protected.HandleFunc("/notes/{note_id}", deliveries.NotesDelivery.GetNote).Methods("GET")
	protected.HandleFunc("/notes/{note_id}", deliveries.NotesDelivery.UpdateNote).Methods("PUT")
	protected.HandleFunc("/notes/{note_id}", deliveries.NotesDelivery.DeleteNote).Methods("DELETE")

	// блоки
	protected.HandleFunc("/notes/{note_id}/blocks", deliveries.BlocksDelivery.CreateBlock).Methods("POST")
	protected.HandleFunc("/notes/{note_id}/blocks", deliveries.BlocksDelivery.GetBlocks).Methods("GET")
	protected.HandleFunc("/blocks/{block_id}", deliveries.BlocksDelivery.DeleteBlock).Methods("DELETE")
	protected.HandleFunc("/blocks/{block_id}/position", deliveries.BlocksDelivery.UpdateBlockPosition).Methods("PUT")

	// текст
	protected.HandleFunc("/blocks/{block_id}/text", deliveries.BlocksDelivery.UpdateBlockText).Methods("PUT")
	protected.HandleFunc("/blocks/{block_id}/text", deliveries.BlocksDelivery.GetBlockText).Methods("GET")

	// форматы
	protected.HandleFunc("/blocks/{block_id}/text/formats/apply", deliveries.BlocksDelivery.ApplyFormatToRange).Methods("POST")
	protected.HandleFunc("/blocks/{block_id}/text/formats/remove", deliveries.BlocksDelivery.RemoveFormatFromRange).Methods("POST")
	protected.HandleFunc("/blocks/{block_id}/text/formats", deliveries.BlocksDelivery.GetTextFormats).Methods("GET")

	return mw.CORS(r)
}
