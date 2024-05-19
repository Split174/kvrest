package api

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"go.etcd.io/bbolt"
)

var usersDbPath = "./data/users.db"

// SetUsersDbPath sets the path for users database
func SetUsersDbPath(path string) {
	usersDbPath = path
}

// CreateKV handles the creation of a new KV store and generates an API key for the user.
func createKV(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db, err := bbolt.Open(usersDbPath, 0666, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var apiKey string
	db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("users"))
		if bucket == nil {
			return nil
		}
		apiKey = string(bucket.Get([]byte(req.Name)))
		return nil
	})

	if apiKey != "" {
		http.Error(w, "Name already exists", http.StatusConflict)
		return
	}

	apiKey, err = generateAPIKey()
	if err != nil {
		http.Error(w, "Failed to generate API key", http.StatusInternalServerError)
		return
	}

	db.Update(func(tx *bbolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("users"))
		if err != nil {
			return err
		}
		return bucket.Put([]byte(req.Name), []byte(apiKey))
	})

	os.OpenFile(filepath.Join(dataPath, apiKey+".db"), os.O_RDONLY|os.O_CREATE, 0666)

	response := map[string]string{"api_key": apiKey}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ChangeApiKey changes the user's API key and renames the database file.
func changeApiKey(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db, err := bbolt.Open(usersDbPath, 0666, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var oldApiKey string
	db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("users"))
		if bucket == nil {
			return bbolt.ErrBucketNotFound
		}
		oldApiKey = string(bucket.Get([]byte(req.Name)))
		return nil
	})

	if oldApiKey == "" {
		http.Error(w, "Name not found", http.StatusNotFound)
		return
	}

	newApiKey, err := generateAPIKey()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("users"))
		return bucket.Put([]byte(req.Name), []byte(newApiKey))
	})

	oldDbFile := filepath.Join(dataPath, oldApiKey+".db")
	newDbFile := filepath.Join(dataPath, newApiKey+".db")
	err = os.Rename(oldDbFile, newDbFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{"api_key": newApiKey}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
