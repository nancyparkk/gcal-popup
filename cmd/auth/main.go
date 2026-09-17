// Command auth re-runs the Google OAuth consent flow and caches a fresh token.
// Run it from a terminal when calendar writes start failing with "invalid_grant".
package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"

	"github.com/nancyparkk/gcal-popup/internal/calendar"
)

func main() {
	// calendar.NewService only starts the consent flow when no token file is readable.
	if err := os.Rename("token.json", "token.json.bak"); err != nil && !errors.Is(err, fs.ErrNotExist) {
		log.Fatalf("Unable to set aside existing token: %v", err)
	}

	if _, err := calendar.NewService(context.Background()); err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}

	fmt.Println("Authenticated. Previous token saved as token.json.bak")
}
