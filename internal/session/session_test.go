package session

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/nancyparkk/gcal-popup/internal/llm"
)

type fakeExtractor struct {
	alwaysAsk bool
	err       error
	calls     [][]string
}

func (f *fakeExtractor) Extract(_ context.Context, conversation []string) (*llm.ExtractedEvent, error) {
	f.calls = append(f.calls, append([]string(nil), conversation...))
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

func TestRounds(t *testing.T) {
	tests := []struct {
		length int
		want   int
	}{
		{0, 0},
		{1, 0}, // just initial text
		{3, 1}, // initial + one Q/A
		{5, 2},
		{7, 3},
		{9, 4},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("length_%d", tt.length), func(t *testing.T) {
			conversation := make([]string, tt.length)
			if got := Rounds(conversation); got != tt.want {
				t.Fatalf("Rounds(len %d) = %d, want %d", tt.length, got, tt.want)
			}
		})
	}
}

func TestNextAllowsClarificationBelowCap(t *testing.T) {
	extractor := &fakeExtractor{alwaysAsk: true}

	event, err := Next(context.Background(), extractor, []string{"lunch"})
	if err != nil {
		t.Fatalf("Next() error = %v", err)
	}
	if !event.NeedsClarification {
		t.Fatal("NeedsClarification = false, want true below the cap")
	}
	if event.ClarifyingQuestion == "" {
		t.Fatal("ClarifyingQuestion is empty, want the model's question preserved")
	}
}

func TestNextClearsClarificationAtCap(t *testing.T) {
	extractor := &fakeExtractor{alwaysAsk: true}

	// initial text and 3 qA pairs
	conversation := []string{"lunch", "Q:", "A:", "Q:", "A:", "Q:", "A:"}
	if got := Rounds(conversation); got != MaxClarificationRounds {
		t.Fatalf("test fixture has %d rounds, want %d", got, MaxClarificationRounds)
	}

	event, err := Next(context.Background(), extractor, conversation)
	if err != nil {
		t.Fatalf("Next() error = %v", err)
	}
	if event.NeedsClarification {
		t.Fatal("NeedsClarification = true at the cap, want it forced to false")
	}
	if event.ClarifyingQuestion != "" {
		t.Fatalf("ClarifyingQuestion = %q at the cap, want it cleared", event.ClarifyingQuestion)
	}
}

func TestRunTerminatesAtCap(t *testing.T) {
	extractor := &fakeExtractor{alwaysAsk: true}
	asked := 0

	event, err := Run(context.Background(), extractor, "lunch", func(string) (string, error) {
		asked++
		return "noon", nil
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if asked != MaxClarificationRounds {
		t.Fatalf("asked %d questions, want %d", asked, MaxClarificationRounds)
	}
	if event.NeedsClarification {
		t.Fatal("Run returned an event still wanting clarification")
	}
}

func TestRunStopsWhenModelIsSatisfied(t *testing.T) {
	extractor := &fakeExtractor{alwaysAsk: false}
	asked := 0

	if _, err := Run(context.Background(), extractor, "lunch tomorrow at noon", func(string) (string, error) {
		asked++
		return "", nil
	}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if asked != 0 {
		t.Fatalf("asked %d questions, want 0 when no clarification is needed", asked)
	}
	if len(extractor.calls) != 1 {
		t.Fatalf("extractor called %d times, want 1", len(extractor.calls))
	}
}

func TestRunAppendsAnswersToConversation(t *testing.T) {
	extractor := &fakeExtractor{alwaysAsk: true}

	if _, err := Run(context.Background(), extractor, "lunch", func(string) (string, error) {
		return "noon", nil
	}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	wantLengths := []int{1, 3, 5, 7}
	if len(extractor.calls) != len(wantLengths) {
		t.Fatalf("extractor called %d times, want %d", len(extractor.calls), len(wantLengths))
	}
	for i, want := range wantLengths {
		if got := len(extractor.calls[i]); got != want {
			t.Fatalf("call %d saw conversation of length %d, want %d", i, got, want)
		}
	}
}

func TestNextPropagatesExtractorError(t *testing.T) {
	wantErr := errors.New("gemini unavailable")
	extractor := &fakeExtractor{err: wantErr}

	if _, err := Next(context.Background(), extractor, []string{"lunch"}); !errors.Is(err, wantErr) {
		t.Fatalf("Next() error = %v, want %v", err, wantErr)
	}
}
