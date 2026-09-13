package llm

import (
	"testing"
	"time"
)

func TestValidate(t *testing.T) {
	valid := func() ExtractedEvent {
		return ExtractedEvent{
			Title:           "lunch with sarah",
			Date:            time.Now().AddDate(0, 0, 7).Format("2006-01-02"),
			StartTime:       "13:00",
			DurationMinutes: 60,
		}
	}

	tests := []struct {
		name    string
		mutate  func(*ExtractedEvent)
		wantErr bool
	}{
		{"unmodified event is valid", func(e *ExtractedEvent) {}, false},

		{"empty title", func(e *ExtractedEvent) { e.Title = "" }, true},
		{"whitespace-only title", func(e *ExtractedEvent) { e.Title = "   " }, true},

		{"empty date", func(e *ExtractedEvent) { e.Date = "" }, true},
		{"unparsed natural language date", func(e *ExtractedEvent) { e.Date = "next tuesday" }, true},
		{"wrong date layout", func(e *ExtractedEvent) { e.Date = "09/17/2026" }, true},
		{"date years in the past", func(e *ExtractedEvent) { e.Date = "2020-01-01" }, true},
		{"date beyond a year out", func(e *ExtractedEvent) {
			e.Date = time.Now().AddDate(1, 0, 7).Format("2006-01-02")
		}, true},
		{"today is allowed", func(e *ExtractedEvent) {
			e.Date = time.Now().Format("2006-01-02")
		}, false},
		{"yesterday is allowed as grace", func(e *ExtractedEvent) {
			e.Date = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
		}, false},

		{"empty start time", func(e *ExtractedEvent) { e.StartTime = "" }, true},
		{"12-hour start time", func(e *ExtractedEvent) { e.StartTime = "1pm" }, true},
		{"hour out of range", func(e *ExtractedEvent) { e.StartTime = "25:00" }, true},
		{"midnight is valid", func(e *ExtractedEvent) { e.StartTime = "00:00" }, false},

		{"zero duration", func(e *ExtractedEvent) { e.DurationMinutes = 0 }, true},
		{"duration below minimum", func(e *ExtractedEvent) { e.DurationMinutes = 4 }, true},
		{"duration at minimum", func(e *ExtractedEvent) { e.DurationMinutes = 5 }, false},
		{"duration at maximum", func(e *ExtractedEvent) { e.DurationMinutes = 720 }, false},
		{"duration above maximum", func(e *ExtractedEvent) { e.DurationMinutes = 721 }, true},
		{"negative duration", func(e *ExtractedEvent) { e.DurationMinutes = -30 }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := valid()
			tt.mutate(&event)

			err := event.Validate()
			if tt.wantErr && err == nil {
				t.Fatalf("Validate() = nil, want an error (event: %+v)", event)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("Validate() = %v, want nil (event: %+v)", err, event)
			}
		})
	}
}
