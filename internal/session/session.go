package session

import (
	"context"
	"fmt"

	"github.com/nancyparkk/gcal-popup/internal/llm"
)

//caps number of questions before fallback to best guess
const maxClarificationRounds = 3

//called when sess needs clarifying question
type AskUserFunc func(question string) (string, error)

//manages one capture
type Session struct {
	extractor    llm.Extractor
	conversation []string
}

//new session with the user's initial freeform text
func New(extractor llm.Extractor, initialText string) *Session {
	return &Session{
		extractor:    extractor,
		conversation: []string{initialText},
	}
}

// executes and returns best event
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

	//return best guess
	return lastEvent, nil
}
