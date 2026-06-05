package main

import (
	"log"
	"net/http"

	"github.com/EOEboh/mb-project-01-chat/handlers"
)

func main() {
	mux := http.NewServeMux()

	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Project 1 routes
	// GET  /     : serve the chat page
	// POST /chat : receive a message, stream the AI response back via SSE
	mux.HandleFunc("/", handlers.Index)
	mux.HandleFunc("POST /chat", handlers.Chat)

	log.Println("🚀 AI Chat running → http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
