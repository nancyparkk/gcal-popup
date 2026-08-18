package calendar

import (
	"context"
	"fmt"
	"time"

	googlecal "google.golang.org/api/calendar/v3"
)

// minimal set of fields needed to create a calendar event
type EventInput struct {
	Title           string
	Date            string
	StartTime       string
	DurationMinutes int
}

// creates an event on the user's primary gcal
func WriteEvent(ctx context.Context, srv *googlecal.Service, input EventInput) (string, error) {
	start, err := time.ParseInLocation("2006-01-02 15:04", input.Date+" "+input.StartTime, time.Local)
	if err != nil {
		return "", fmt.Errorf("invalid date/time: %w", err)
	}
	end := start.Add(time.Duration(input.DurationMinutes) * time.Minute)

	event := &googlecal.Event{
		Summary: input.Title,
		Start:   &googlecal.EventDateTime{DateTime: start.Format(time.RFC3339)},
		End:     &googlecal.EventDateTime{DateTime: end.Format(time.RFC3339)},
	}

	created, err := srv.Events.Insert("primary", event).Do()
	if err != nil {
		return "", fmt.Errorf("unable to create event: %w", err)
	}

	return created.HtmlLink, nil
}
