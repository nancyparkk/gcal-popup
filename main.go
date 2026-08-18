package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/nancyparkk/gcal-popup/internal/calendar"
	googlecal "google.golang.org/api/calendar/v3"
)

func main() {
	ctx := context.Background()

	srv, err := calendar.NewService(ctx)
	if err != nil {
		log.Fatalf("Unable to create calendar service: %v", err)
	}

	start := time.Now().Add(1 * time.Hour)
	end := start.Add(30 * time.Minute)

	event := &googlecal.Event{
		Summary: "gcal-popup test event",
		Start: &googlecal.EventDateTime{
			DateTime: start.Format(time.RFC3339),
		},
		End: &googlecal.EventDateTime{
			DateTime: end.Format(time.RFC3339),
		},
	}

	created, err := srv.Events.Insert("primary", event).Do()
	if err != nil {
		log.Fatalf("Unable to create event: %v", err)
	}

	fmt.Printf("Event created: %s\n", created.HtmlLink)
}
