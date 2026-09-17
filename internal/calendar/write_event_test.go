package calendar

import (
	"testing"
	"time"
)

func TestEventTimes(t *testing.T) {
	tests := []struct {
		name           string
		input          EventInput
		wantStart      string
		wantDurMin     float64
		wantErr        bool
		wantCrossesDay bool
	}{
		{
			name:       "afternoon meeting",
			input:      EventInput{Date: "2026-09-17", StartTime: "13:00", DurationMinutes: 90},
			wantStart:  "2026-09-17T13:00:00",
			wantDurMin: 90,
		},
		{
			name:       "midnight start",
			input:      EventInput{Date: "2026-09-17", StartTime: "00:00", DurationMinutes: 30},
			wantStart:  "2026-09-17T00:00:00",
			wantDurMin: 30,
		},
		{
			name:           "duration rolls past midnight",
			input:          EventInput{Date: "2026-09-17", StartTime: "23:30", DurationMinutes: 60},
			wantStart:      "2026-09-17T23:30:00",
			wantDurMin:     60,
			wantCrossesDay: true,
		},
		{
			name:    "malformed date",
			input:   EventInput{Date: "next tuesday", StartTime: "13:00", DurationMinutes: 30},
			wantErr: true,
		},
		{
			name:    "malformed time",
			input:   EventInput{Date: "2026-09-17", StartTime: "1pm", DurationMinutes: 30},
			wantErr: true,
		},
		{
			name:    "empty input",
			input:   EventInput{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := eventTimes(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatal("eventTimes() = nil error, want an error")
				}
				return
			}
			if err != nil {
				t.Fatalf("eventTimes() error = %v", err)
			}

			wantStart, perr := time.ParseInLocation("2006-01-02T15:04:05", tt.wantStart, time.Local)
			if perr != nil {
				t.Fatalf("bad fixture: %v", perr)
			}
			if !start.Equal(wantStart) {
				t.Fatalf("start = %v, want %v", start, wantStart)
			}
			if got := end.Sub(start).Minutes(); got != tt.wantDurMin {
				t.Fatalf("duration = %v minutes, want %v", got, tt.wantDurMin)
			}
			if crossed := end.Day() != start.Day(); crossed != tt.wantCrossesDay {
				t.Fatalf("crosses day boundary = %v, want %v", crossed, tt.wantCrossesDay)
			}
		})
	}
}

func TestEventTimesUsesLocalZone(t *testing.T) {
	start, _, err := eventTimes(EventInput{Date: "2026-09-17", StartTime: "13:00", DurationMinutes: 30})
	if err != nil {
		t.Fatalf("eventTimes() error = %v", err)
	}
	if start.Location() != time.Local {
		t.Fatalf("start zone = %v, want %v", start.Location(), time.Local)
	}
	if h := start.Hour(); h != 13 {
		t.Fatalf("local hour = %d, want 13", h)
	}
}
