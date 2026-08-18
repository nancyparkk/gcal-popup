package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/nancyparkk/gcal-popup/internal/calendar"
	"github.com/nancyparkk/gcal-popup/internal/llm"
	"github.com/nancyparkk/gcal-popup/internal/session"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on existing environment variables")
	}

	ctx := context.Background()

	extractor, err := llm.NewGeminiExtractor(ctx)
	if err != nil {
		log.Fatalf("Unable to create extractor: %v", err)
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("What's the event? ")
	initialText, _ := reader.ReadString('\n')
	initialText = strings.TrimSpace(initialText)

	sess := session.New(extractor, initialText)

	askUser := func(question string) (string, error) {
		fmt.Printf("\n%s\n> ", question)
		answer, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(answer), nil
	}

	event, err := sess.Run(ctx, askUser)
	if err != nil {
		log.Fatalf("Session failed: %v", err)
	}

	fmt.Printf("\nProposed event\n%+v\n", event)

	if err := event.Validate(); err != nil {
		log.Fatalf("Event failed validation, refusing to write: %v", err)
	}

	fmt.Print("\nCreate this event? (y/n) ")
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm != "y" {
		fmt.Println("Cancelled.")
		return
	}

	srv, err := calendar.NewService(ctx)
	if err != nil {
		log.Fatalf("Unable to create calendar service: %v", err)
	}

	link, err := calendar.WriteEvent(ctx, srv, calendar.EventInput{
		Title:           event.Title,
		Date:            event.Date,
		StartTime:       event.StartTime,
		DurationMinutes: event.DurationMinutes,
	})
	if err != nil {
		log.Fatalf("Unable to write event: %v", err)
	}

	fmt.Printf("\nEvent created: %s\n", link)
}
