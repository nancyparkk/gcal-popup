// Package session owns the clarification policy: how many follow-up questions
// the model may ask before it has to commit to a best guess.
package session

import (
	"context"

	"github.com/nancyparkk/gcal-popup/internal/llm"
)

// MaxClarificationRounds caps follow-up questions before falling back to a best guess.
const MaxClarificationRounds = 3

// AskUserFunc collects an answer to a clarifying question.
type AskUserFunc func(question string) (string, error)

// Rounds reports how many question/answer exchanges a conversation already holds.
// A conversation is the initial text followed by two entries per round.
func Rounds(conversation []string) int {
	if len(conversation) < 2 {
		return 0
	}
	return (len(conversation) - 1) / 2
}

// Next performs one extraction and applies the clarification cap. Once the cap is
// reached the event comes back final no matter what the model asked for, so no
// caller can loop forever.
func Next(ctx context.Context, extractor llm.Extractor, conversation []string) (*llm.ExtractedEvent, error) {
	event, err := extractor.Extract(ctx, conversation)
	if err != nil {
		return nil, err
	}

	if Rounds(conversation) >= MaxClarificationRounds {
		event.NeedsClarification = false
		event.ClarifyingQuestion = ""
	}
	return event, nil
}

// Run drives the clarification loop to completion for terminal front ends, which
// can block waiting on input. HTTP callers use Next directly and keep the
// conversation on the client instead.
func Run(ctx context.Context, extractor llm.Extractor, initialText string, ask AskUserFunc) (*llm.ExtractedEvent, error) {
	conversation := []string{initialText}

	for {
		event, err := Next(ctx, extractor, conversation)
		if err != nil {
			return nil, err
		}
		if !event.NeedsClarification {
			return event, nil
		}

		answer, err := ask(event.ClarifyingQuestion)
		if err != nil {
			return nil, err
		}
		conversation = append(conversation, "Q: "+event.ClarifyingQuestion, "A: "+answer)
	}
}
