package server

import (
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/nancyparkk/gcal-popup/internal/calendar"
	"github.com/nancyparkk/gcal-popup/internal/llm"
	"github.com/nancyparkk/gcal-popup/internal/session"
)

//go:embed static
var staticFiles embed.FS

type extractRequest struct {
	Conversation []string `json:"conversation"`
}

type createRequest struct {
	Title           string `json:"title"`
	Date            string `json:"date"`
	StartTime       string `json:"start_time"`
	DurationMinutes int    `json:"duration_minutes"`
}

type createResponse struct {
	Link string `json:"link"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type Server struct {
	extractor llm.Extractor
}

func New(extractor llm.Extractor) *Server {
	return &Server{extractor: extractor}
}

func (s *Server) Handler() http.Handler {
	static, err := fs.Sub(staticFiles, "static")
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/extract", s.handleExtract)
	mux.HandleFunc("POST /api/create", s.handleCreate)
	mux.Handle("GET /", http.FileServer(http.FS(static)))
	return mux
}

func Start(addr string) error {
	_ = godotenv.Load()

	extractor, err := llm.NewGeminiExtractor(context.Background())
	if err != nil {
		return err
	}

	return http.ListenAndServe(addr, New(extractor).Handler())
}

func (s *Server) handleExtract(w http.ResponseWriter, r *http.Request) {
	var req extractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	if len(req.Conversation) == 0 {
		writeError(w, http.StatusBadRequest, "conversation is empty")
		return
	}

	event, err := session.Next(r.Context(), s.extractor, req.Conversation)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, event)
}

func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	event := &llm.ExtractedEvent{
		Title:           req.Title,
		Date:            req.Date,
		StartTime:       req.StartTime,
		DurationMinutes: req.DurationMinutes,
	}
	if err := event.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	srv, err := calendar.NewService(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	link, err := calendar.WriteEvent(r.Context(), srv, calendar.EventInput{
		Title:           req.Title,
		Date:            req.Date,
		StartTime:       req.StartTime,
		DurationMinutes: req.DurationMinutes,
	})
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, createResponse{Link: link})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}
