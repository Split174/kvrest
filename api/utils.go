package api

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

func RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/{bucketName}", createBucket).Methods("PUT")
	r.HandleFunc("/{bucketName}/{key}", setKey).Methods("PUT")
	r.HandleFunc("/{bucketName}/{key}", getValue).Methods("GET")
	r.HandleFunc("/{bucketName}/{key}", deleteKey).Methods("DELETE")
}

func RegisterAdminRoutes(r *mux.Router) {
	r.HandleFunc("/create_kv", createKV).Methods("PUT")
	r.HandleFunc("/change_api_key", changeApiKey).Methods("PUT")
}

func ApiKeyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("API-Key")
		if apiKey == "" {
			http.Error(w, "Missing API key", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("Started %s %s for %s", r.Method, r.RequestURI, r.RemoteAddr)

		next.ServeHTTP(w, r)

		log.Printf("Completed %s %s in %v", r.Method, r.RequestURI, time.Since(start))
	})
}

func MasterKeyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		masterKey := r.Header.Get("MASTER-API-KEY")
		if masterKey == "" || masterKey != os.Getenv("MASTER_API_KEY") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func generateAPIKey() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
