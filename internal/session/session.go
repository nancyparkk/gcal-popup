package session

import (
	"context"
	"fmt"

	"github.com/nancyparkk/gcal-popup/internal/llm"
)

// maxClarificationRounds caps how many times the session will ask a
// clarifying question before falling back to its best guess.
const maxClarificationRounds = 3

// AskUserFunc is called when the session needs to ask a clarifying question.
// It's given the question and should return the user's answer. This keeps
// Session decoupled from any specific UI (terminal, webview, etc).
type AskUserFunc func(question string) (string, error)

// Session manages one capture: an initial piece of text, run through the
// LLM extractor, looping through clarifying questions (bounded) until
// confident or out of rounds.
type Session struct {
	extractor    llm.Extractor
	conversation []string
}

// New starts a new session with the user's initial freeform text.
func New(extractor llm.Extractor, initialText string) *Session {
	return &Session{
		extractor:    extractor,
		conversation: []string{initialText},
	}
}

// Run executes the extract -> clarify loop, capped at maxClarificationRounds,
// and returns the final best-effort extracted event.
func (s *Session) Run(ctx context.Context, askUser AskUserFunc) (*llm.ExtractedEvent, error) {
	var lastEvent *llm.ExtractedEvent

	for round := 0; round < maxClarificationRounds; round++ {
		event, err := s.extractor.Extract(ctx, s.conversation)
		if err != nil {
			return nil, fmt.Errorf("extraction failed on round %d: %w", round+1, err)
		}
		lastEvent = event

		if !event.NeedsClarification {
			return event, nil
		}

		answer, err := askUser(event.ClarifyingQuestion)
		if err != nil {
			return nil, fmt.Errorf("failed to get user answer: %w", err)
		}

		s.conversation = append(s.conversation, "Q: "+event.ClarifyingQuestion, "A: "+answer)
	}

	// Out of rounds - return the best guess we have, even though it may
	// still be flagged as needing clarification. The caller (UI/confirmation
	// step) is responsible for letting the user manually correct it.
	return lastEvent, nil
}
