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

	// Define the main router
	router := mux.NewRouter()

	// Define the subrouter for API with both middlewares
	apiRouter := router.PathPrefix("/api").Subrouter()
	api.RegisterRoutes(apiRouter)
	apiRouter.Use(api.ApiKeyMiddleware)
	apiRouter.Use(api.DisableSystemBucketMiddleware)
	// Apply Logging Middleware to all routes
	router.Use(api.LoggingMiddleware)

	log.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
