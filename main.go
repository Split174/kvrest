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

	// Define the subrouter for admin with master key middleware
	adminRouter := router.PathPrefix("/admin").Subrouter()
	RegisterAdminRoutes(adminRouter)
	adminRouter.Use(api.MasterKeyMiddleware)

	// Define the subrouter for API with both middlewares
	apiRouter := router.PathPrefix("/api").Subrouter()
	RegisterRoutes(apiRouter)
	apiRouter.Use(api.ApiKeyMiddleware)

	// Apply Logging Middleware to all routes
	router.Use(api.LoggingMiddleware)

	log.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/{bucketName}", api.CreateBucket).Methods("PUT")
	r.HandleFunc("/{bucketName}/{key}", api.SetKey).Methods("PUT")
	r.HandleFunc("/{bucketName}/{key}", api.GetValue).Methods("GET")
	r.HandleFunc("/{bucketName}/{key}", api.DeleteKey).Methods("DELETE")
}

func RegisterAdminRoutes(r *mux.Router) {
	r.HandleFunc("/create_kv", api.CreateKV).Methods("PUT")
	r.HandleFunc("/change_api_key", api.ChangeApiKey).Methods("PUT")
}
