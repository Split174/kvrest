package main

import (
	"kvest/api"
	"kvest/telegram_bot"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	// Ensure the /data/ directory exists
	dataDir := "./data/"
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		log.Println("Creating /data/ directory...")
		err := os.MkdirAll(dataDir, 0755)
		if err != nil {
			log.Fatalf("Failed to create /data/ directory: %s", err)
		}
	}

	// Start the Telegram bot in a separate goroutine
	go telegram_bot.StartBot()

	// Define the API router
	router := mux.NewRouter()
	api.RegisterRoutes(router)

	// Apply middlewares
	router.Use(api.LoggingMiddleware)
	router.Use(api.ApiKeyMiddleware)

	// Apply master key middleware only to admin routes
	adminRouter := router.PathPrefix("/admin").Subrouter()
	adminRouter.Use(api.MasterKeyMiddleware)

	log.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
