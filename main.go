package main

import (
	"backend/handler"
	"backend/store"
	"errors"
	"log"
	"net/http"
)

const host = "localhost"
const port = "8080"

func main() {
	s := store.NewStore()

	s.InitFillStore()

	router := handler.NewRouter(s)

	server := &http.Server{
		Addr:    host + ":" + port,
		Handler: router,
	}

	log.Printf("listening on %s\n", server.Addr)
	err := server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server error: %v", err)
	}
}
