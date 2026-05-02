package main

import (
	"log"
	"net/http"
	"os"

	"github.com/ajedrez4/ajedrez4-server/internal/handler"
	"github.com/ajedrez4/ajedrez4-server/internal/store"
)

func main() {
	sessionStore := store.NewSessionStore()
	gameStore := store.NewGameStore()

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, sessionStore, gameStore)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("starting server on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
