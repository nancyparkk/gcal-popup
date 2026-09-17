package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nancyparkk/gcal-popup/internal/llm"
)

type fakeExtractor struct {
	alwaysAsk bool
	err       error
}

func (f *fakeExtractor) Extract(_ context.Context, _ []string) (*llm.ExtractedEvent, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &llm.ExtractedEvent{
		Title:              "Lunch",
		Date:               "2026-09-17",
		StartTime:          "13:00",
		DurationMinutes:    60,
		NeedsClarification: f.alwaysAsk,
		ClarifyingQuestion: "What time?",
	}, nil
}

func post(t *testing.T, h http.Handler, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func decodeEvent(t *testing.T, rec *httptest.ResponseRecorder) llm.ExtractedEvent {
	t.Helper()
	var event llm.ExtractedEvent
	if err := json.Unmarshal(rec.Body.Bytes(), &event); err != nil {
		t.Fatalf("response is not a valid event: %v (body: %s)", err, rec.Body)
	}
	return event
}

func decodeError(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	var resp errorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not a valid error: %v (body: %s)", err, rec.Body)
	}
	return resp.Error
}

func TestExtractReturnsEvent(t *testing.T) {
	h := New(&fakeExtractor{}).Handler()

	rec := post(t, h, "/api/extract", `{"conversation":["lunch tomorrow at noon"]}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body)
	}
	if event := decodeEvent(t, rec); event.Title != "Lunch" {
		t.Fatalf("title = %q, want %q", event.Title, "Lunch")
	}
}

func TestExtractPassesThroughClarification(t *testing.T) {
	h := New(&fakeExtractor{alwaysAsk: true}).Handler()

	rec := post(t, h, "/api/extract", `{"conversation":["lunch"]}`)
	if event := decodeEvent(t, rec); !event.NeedsClarification {
		t.Fatal("NeedsClarification = false, want true below the cap")
	}
}

func TestExtractEnforcesClarificationCap(t *testing.T) {
	h := New(&fakeExtractor{alwaysAsk: true}).Handler()

	rec := post(t, h, "/api/extract",
		`{"conversation":["lunch","Q:","A:","Q:","A:","Q:","A:"]}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	event := decodeEvent(t, rec)
	if event.NeedsClarification {
		t.Fatal("NeedsClarification = true past the cap, want it forced false")
	}
	if event.ClarifyingQuestion != "" {
		t.Fatalf("ClarifyingQuestion = %q past the cap, want it cleared", event.ClarifyingQuestion)
	}
}

func TestExtractRejectsEmptyConversation(t *testing.T) {
	h := New(&fakeExtractor{}).Handler()

	rec := post(t, h, "/api/extract", `{"conversation":[]}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if msg := decodeError(t, rec); msg != "conversation is empty" {
		t.Fatalf("error = %q, want %q", msg, "conversation is empty")
	}
}

func TestExtractRejectsMalformedJSON(t *testing.T) {
	h := New(&fakeExtractor{}).Handler()

	if rec := post(t, h, "/api/extract", `{"conversation":`); rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestExtractReportsUpstreamFailureAsBadGateway(t *testing.T) {
	h := New(&fakeExtractor{err: errors.New("gemini unavailable")}).Handler()

	if rec := post(t, h, "/api/extract", `{"conversation":["lunch"]}`); rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadGateway)
	}
}

func TestCreateRejectsInvalidEventBeforeCallingGoogle(t *testing.T) {
	h := New(&fakeExtractor{}).Handler()

	rec := post(t, h, "/api/create",
		`{"title":"","date":"2020-01-01","start_time":"99:99","duration_minutes":0}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusBadRequest, rec.Body)
	}
	if msg := decodeError(t, rec); msg != "title is empty" {
		t.Fatalf("error = %q, want %q", msg, "title is empty")
	}
}

func TestServesFrontend(t *testing.T) {
	h := New(&fakeExtractor{}).Handler()

	for _, path := range []string{"/", "/app.js", "/style.css"} {
		t.Run(path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
			}
			if rec.Body.Len() == 0 {
				t.Fatal("empty body")
			}
		})
	}
}
