package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
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

	fmt.Printf("\n--- Final result ---\n%+v\n", event)
}
