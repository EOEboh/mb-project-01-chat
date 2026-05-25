# Project 01 — AI Chat Interface

> **Bootcamp Day 3** · Tier 1 (Foundation) · Est. build time: 60–90 min

Stream real-time LLM responses to a browser using Go, Ollama, and Server-Sent Events.

This is the "Hello World" of AI backends — every concept introduced here appears
in every project that follows.

---

## What You'll Build

A single-page chat application where:
- The user types a message and hits Enter
- The browser **POSTs** it to `/chat`
- Go opens an SSE connection and starts **streaming tokens** from Ollama
- Each token appears in the browser **as it's generated** — no waiting for the full response

---

## Prerequisites

- [Go 1.22+](https://go.dev/dl/)
- [Ollama](https://ollama.com) installed and running

```bash
# Install the model (one-time, ~2GB download)
ollama pull llama3.2:3b

# Verify Ollama is running
ollama serve
```

---

## Run It

```bash
git clone https://github.com/YOUR_GITHUB_USERNAME/project-01-chat
cd project-01-chat
go run main.go
# => http://localhost:8080
```

---

## Architecture

```
Browser                    Go Server                    Ollama
  │                            │                            │
  │  POST /chat                │                            │
  │  body: message=Hello       │                            │
  ├──────────────────────────► │                            │
  │                            │  POST /api/chat            │
  │                            │  stream: true              │
  │                            ├──────────────────────────► │
  │                            │                            │
  │  HTTP 200                  │  {"message":{"content":"H"}│
  │  Content-Type:             │ ◄──────────────────────────┤
  │    text/event-stream       │  {"message":{"content":"i"}│
  │ ◄──────────────────────────┤ ◄──────────────────────────┤
  │                            │  {"done": true}            │
  │  data: "H"\n\n             │ ◄──────────────────────────┤
  │ ◄──────────────────────────┤                            │
  │  data: "i"\n\n             │                            │
  │ ◄──────────────────────────┤                            │
  │  data: "[DONE]"\n\n        │                            │
  │ ◄──────────────────────────┤                            │
```

---

## Key Files

| File | Responsibility |
|------|----------------|
| `main.go` | Register routes, start server |
| `handlers/chat.go` | SSE connection setup + stream loop |
| `ai/ollama.go` | All Ollama communication — `Chat()` and `ChatStream()` |
| `templates/index.html` | Chat UI + JavaScript stream reader |

---

## Key Concepts

### Server-Sent Events (SSE)
Three headers turn a normal HTTP response into a live stream:
```go
w.Header().Set("Content-Type", "text/event-stream")
w.Header().Set("Cache-Control", "no-cache")
w.Header().Set("Connection", "keep-alive")
```
Each event is a line starting with `data:` followed by two newlines:
```
data: "Hello"\n\n
data: " world"\n\n
data: "[DONE]"\n\n
```

### Why We JSON-Encode Each Chunk
Ollama responses can contain `\n`, `"`, and other characters that would break raw SSE.
Wrapping each chunk in `json.Marshal` makes it safe:
```go
encoded, _ := json.Marshal(chunk)
fmt.Fprintf(w, "data: %s\n\n", encoded)
```
The browser unwraps it with `JSON.parse(payload)`.

### Why fetch() Instead of EventSource
`EventSource` only supports GET requests. Since we POST the user message, we use
`fetch()` + `ReadableStream` instead. The reading logic is identical — just more explicit.

---

## Off-Day Extension (Async Task)

Pick one:
1. **Add conversation history** — store previous turns and include them in every Ollama request so the AI remembers context
2. **Add a system prompt editor** — let the user change the AI's persona via a settings panel
3. **Add a model switcher** — a dropdown that switches between `llama3.2:3b` and `phi3:mini`

These patterns appear in Projects 5, 7, and 10 — building them now gives you a head start.

---

## What's Next

→ **Project 02: Code Snippet Explainer** — same SSE pattern, but we introduce
  code-specific models, system prompt engineering, and output parsing.