package main

import (
	"kvest/api"
	"kvest/telegram_bot"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	// Start the Telegram bot in a separate goroutine
	go telegram_bot.StartBot()

	// Define the API router
	router := mux.NewRouter()
	api.RegisterRoutes(router)
	router.Use(api.LoggingMiddleware)
	router.Use(api.ApiKeyMiddleware)

	log.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
