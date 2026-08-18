package main

import (
	"context"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/nancyparkk/gcal-popup/internal/llm"
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

	testInputs := []string{
		"lunch with sarah thursday around 1ish",
		"dentist appointment next tuesday 9am",
		"finish essay by friday",
	}

	for _, input := range testInputs {
		fmt.Printf("\n--- Input: %q ---\n", input)

		conversation := []string{input}
		event, err := extractor.Extract(ctx, conversation)
		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}

		fmt.Printf("%+v\n", event)

		if event.NeedsClarification {
			fmt.Printf("Would ask: %s\n", event.ClarifyingQuestion)
		}
	}
}
