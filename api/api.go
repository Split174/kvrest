package api

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	"go.etcd.io/bbolt"
)

var dataPath = "./data/"

const reservedBucket = "kvrest-system-internal"

func createBucket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucketName := vars["bucketName"]

	db, err := openDb(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		return err
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func setKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucketName := vars["bucketName"]
	key := vars["key"]

	var value map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&value); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	valueBytes, err := json.Marshal(value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	db, err := openDb(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	err = db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return bbolt.ErrBucketNotFound
		}
		return bucket.Put([]byte(key), valueBytes)
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func getValue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucketName := vars["bucketName"]
	key := vars["key"]

	db, err := openDb(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var value []byte
	err = db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return bbolt.ErrBucketNotFound
		}
		value = bucket.Get([]byte(key))
		return nil
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if value == nil {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(value)
}

func deleteKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucketName := vars["bucketName"]
	key := vars["key"]

	db, err := openDb(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	err = db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return bbolt.ErrBucketNotFound
		}
		return bucket.Delete([]byte(key))
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func deleteBucket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucketName := vars["bucketName"]

	db, err := openDb(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	err = db.Update(func(tx *bbolt.Tx) error {
		return tx.DeleteBucket([]byte(bucketName))
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func listBuckets(w http.ResponseWriter, r *http.Request) {
	db, err := openDb(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var buckets []string
	err = db.View(func(tx *bbolt.Tx) error {
		return tx.ForEach(func(bucketName []byte, _ *bbolt.Bucket) error {
			if string(bucketName) != reservedBucket {
				buckets = append(buckets, string(bucketName))
			}
			return nil
		})
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]string{"buckets": buckets})
}

func listKeys(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucketName := vars["bucketName"]

	db, err := openDb(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var keys []string
	err = db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return bbolt.ErrBucketNotFound
		}
		return bucket.ForEach(func(k, v []byte) error {
			keys = append(keys, string(k))
			return nil
		})
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]string{"keys": keys})
}

func RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/{bucketName}", createBucket).Methods("PUT")
	r.HandleFunc("/{bucketName}", deleteBucket).Methods("DELETE")
	r.HandleFunc("/buckets", listBuckets).Methods("HEAD")
	r.HandleFunc("/{bucketName}/{key}", setKey).Methods("PUT")
	r.HandleFunc("/{bucketName}/{key}", getValue).Methods("GET")
	r.HandleFunc("/{bucketName}/{key}", deleteKey).Methods("DELETE")
	r.HandleFunc("/{bucketName}", listKeys).Methods("GET")
}

func ApiKeyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("API-KEY")
		if apiKey == "" {
			http.Error(w, "Missing API key", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func DisableSystemBucketMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		bucketName := vars["bucketName"]

		if bucketName == reservedBucket {
			http.Error(w, "Bucket name 'system' not allowed", http.StatusMethodNotAllowed)
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

func openDb(r *http.Request) (*bbolt.DB, error) {
	apiKey := r.Header.Get("API-KEY")
	dbFile := filepath.Join(dataPath, apiKey+".db")

	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return nil, err
	}

	return bbolt.Open(dbFile, 0666, nil)
}
