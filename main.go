package main

import (
	"log"
	"os/exec"

	"github.com/getlantern/systray"
	"github.com/joho/godotenv"
	"github.com/nancyparkk/gcal-popup/internal/server"
)

const addr = "localhost:3000"

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on existing environment variables")
	}

	go func() {
		if err := server.Start(addr); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	systray.Run(onReady, func() {})
}

func onReady() {
	systray.SetTitle("📅")
	systray.SetTooltip("gcal-popup")

	newEvent := systray.AddMenuItem("New event...", "Capture a new calendar event")
	systray.AddSeparator()
	quit := systray.AddMenuItem("Quit", "Quit gcal-popup")

	go func() {
		for {
			select {
			case <-newEvent.ClickedCh:
				if err := exec.Command("open", "http://"+addr).Start(); err != nil {
					log.Printf("Unable to open browser: %v", err)
				}
			case <-quit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}
