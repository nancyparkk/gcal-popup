package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"google.golang.org/genai"
)

// ExtractedEvent is the structured result of parsing freeform text.
type ExtractedEvent struct {
	Title              string `json:"title"`
	Date               string `json:"date"`       // YYYY-MM-DD
	StartTime          string `json:"start_time"` // HH:MM (24hr)
	DurationMinutes    int    `json:"duration_minutes"`
	Confidence         string `json:"confidence"` // high | medium | low
	NeedsClarification bool   `json:"needs_clarification"`
	ClarifyingQuestion string `json:"clarifying_question"`
}

// Extractor turns a conversation (original text + any Q&A so far) into a structured event.
type Extractor interface {
	Extract(ctx context.Context, conversation []string) (*ExtractedEvent, error)
}

// GeminiExtractor implements Extractor using Google's Gemini API.
type GeminiExtractor struct {
	client *genai.Client
	model  string
}

// NewGeminiExtractor creates an Extractor backed by Gemini, using GEMINI_API_KEY from the environment.
func NewGeminiExtractor(ctx context.Context) (*GeminiExtractor, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is not set")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create gemini client: %w", err)
	}

	return &GeminiExtractor{client: client, model: "gemini-3.6-flash"}, nil
}

func (g *GeminiExtractor) Extract(ctx context.Context, conversation []string) (*ExtractedEvent, error) {
	now := time.Now()
	systemPrompt := fmt.Sprintf(`You extract calendar event details from freeform text.

Today's date is %s. The user's timezone is local (assume their machine's local time).

Given the user's message (and any follow-up answers), extract: title, date, start_time, duration_minutes.

If something essential is genuinely ambiguous (you cannot make a reasonable default guess), set needs_clarification to true and ask ONE short, specific clarifying_question. Otherwise set needs_clarification to false and clarifying_question to "".

Respond with ONLY valid JSON, no markdown formatting, matching this exact shape:
{
  "title": "string",
  "date": "YYYY-MM-DD",
  "start_time": "HH:MM",
  "duration_minutes": 30,
  "confidence": "high" | "medium" | "low",
  "needs_clarification": true | false,
  "clarifying_question": "string"
}`, now.Format("2006-01-02 (Monday)"))

	fullPrompt := systemPrompt + "\n\nConversation so far:\n" + strings.Join(conversation, "\n")

	result, err := g.client.Models.GenerateContent(ctx, g.model, genai.Text(fullPrompt), nil)
	if err != nil {
		return nil, fmt.Errorf("gemini request failed: %w", err)
	}

	raw := result.Text()
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var event ExtractedEvent
	if err := json.Unmarshal([]byte(raw), &event); err != nil {
		return nil, fmt.Errorf("failed to parse model response as JSON: %w\nraw response: %s", err, raw)
	}

	return &event, nil
}
