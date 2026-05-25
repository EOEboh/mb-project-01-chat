package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/EOEboh/mb-project-01-chat/ai"
)

var tmpl = template.Must(template.ParseFiles("templates/index.html"))

// Index serves the chat page.
func Index(w http.ResponseWriter, r *http.Request) {
	if err := tmpl.Execute(w, nil); err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

// Chat is the SSE streaming endpoint.
//
// How it works, end to end:
//
//  1. Browser submits a form (POST /chat) with the user's message
//  2. This handler sets SSE headers: these tell the browser:
//     "keep this connection open, I'll push data over it."
//  3. We call ai.ChatStream(), which talks to Ollama with stream:true
//  4. For each token Ollama returns, we write an SSE event to the browser
//  5. The browser reads each event and appends the token to the chat UI
//  6. When Ollama is done, we send a final [DONE] event and close the connection
//
// The browser never reloads. The connection stays alive for the duration
// of the response, then closes. On the next message, a new connection opens.
func Chat(w http.ResponseWriter, r *http.Request) {
	// - 1. Read the user's message from the form body ──────────────────────
	userMessage := r.FormValue("message")
	if userMessage == "" {
		http.Error(w, "message is required", http.StatusBadRequest)
		return
	}

	// - 2. Set SSE headers ─────────────────────────────────────────────────
	//
	// These three headers are the minimum required for SSE:
	//   Content-Type: text/event-stream: tells browser this is an event stream
	//   Cache-Control: no-cache        : don't buffer or cache the stream
	//   Connection: keep-alive         : keep the TCP connection open
	//
	// X-Accel-Buffering: no is for Nginx: without it, Nginx buffers the
	// entire response before forwarding, which breaks streaming entirely.
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	// - 3. Get the Flusher ─────────────────────────────────────────────────
	//
	// http.Flusher lets us push each chunk to the browser immediately.
	// Without Flush(), Go buffers the response and the browser only sees
	// everything at once — the opposite of streaming.
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported by this server", http.StatusInternalServerError)
		return
	}

	// - 4. Build the messages slice ────────────────────────────────────────
	//
	// The system prompt defines the AI's persona and behavior.
	// In later projects (Writing Assistant, AI Tutor) this will be dynamic,
	// loaded from a database per user. For now, it's hardcoded.
	messages := []ai.Message{
		{
			Role:    "system",
			Content: "You are a helpful and concise assistant. Answer clearly and directly.",
		},
		{
			Role:    "user",
			Content: userMessage,
		},
	}

	// - 5. Stream tokens to the browser ────────────────────────────────────
	//
	// SSE event format (the spec):
	//
	//   data: <payload>\n\n
	//
	// Each event MUST end with two newlines. The browser's EventSource API
	// fires an "message" event every time it sees this double-newline.
	//
	// We JSON-encode the chunk so special characters (newlines, quotes) in
	// the AI's response don't break the SSE format.
	err := ai.ChatStream(ai.DefaultModel, messages, func(chunk string) error {
		encoded, _ := json.Marshal(chunk) // e.g. chunk="Hello\nworld" => `"Hello\nworld"`
		fmt.Fprintf(w, "data: %s\n\n", encoded)
		flusher.Flush()
		return nil
	})

	if err != nil {
		log.Printf("stream error: %v", err)
		fmt.Fprintf(w, "data: %s\n\n", `"[Error: AI stream interrupted]"`)
		flusher.Flush()
		return
	}

	// - 6. Signal completion ───────────────────────────────────────────────
	//
	// [DONE] is a sentinel value. The browser checks for this string and
	// stops reading the stream when it sees it.
	fmt.Fprintf(w, "data: \"[DONE]\"\n\n")
	flusher.Flush()
}
