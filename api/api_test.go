package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gorilla/mux"
)

var tempDir string
var apiKey string
var userRouter = mux.NewRouter()

// setupDatabase creates a temporary directory for the database files during testing.
func setupDatabase() error {
	var err error
	tempDir, err = os.MkdirTemp("", "testdb")
	if err != nil {
		return err
	}

	dataPath = tempDir
	apiKey = "test-api-key"
	os.OpenFile(filepath.Join(tempDir, fmt.Sprintf("%s.db", apiKey)), os.O_RDONLY|os.O_CREATE, 0666)
	RegisterRoutes(userRouter)
	userRouter.Use(DisableSystemBucketMiddleware)
	return nil
}

// teardownDatabase removes the temporary directory after testing.
func teardownDatabase() {
	os.RemoveAll(tempDir)
}

func TestE2E(t *testing.T) {
	err := setupDatabase()
	if err != nil {
		t.Fatalf("Could not create temp directory: %v", err)
	}
	defer teardownDatabase()

	// Create a bucket
	req := httptest.NewRequest("PUT", "/testbucket", nil)
	req.Header.Set("API-KEY", apiKey)
	w := httptest.NewRecorder()
	userRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to create bucket: %v", w.Body.String())
	}

	// Set a key-value pair
	value := map[string]interface{}{"name": "test"}
	valueBytes, _ := json.Marshal(value)
	req = httptest.NewRequest("PUT", "/testbucket/testkey", bytes.NewReader(valueBytes))
	req.Header.Set("API-KEY", apiKey)
	w = httptest.NewRecorder()
	userRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to set key: %v", w.Body.String())
	}

	// Get the value
	req = httptest.NewRequest("GET", "/testbucket/testkey", nil)
	req.Header.Set("API-KEY", apiKey)
	w = httptest.NewRecorder()
	userRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to get key: %v", w.Body.String())
	}

	var getValueResponse map[string]interface{}
	json.NewDecoder(w.Body).Decode(&getValueResponse)
	if getValueResponse["name"] != "test" {
		t.Fatalf("Expected 'test', but got %v", getValueResponse["name"])
	}

	// Delete the key
	req = httptest.NewRequest("DELETE", "/testbucket/testkey", nil)
	req.Header.Set("API-KEY", apiKey)
	w = httptest.NewRecorder()
	userRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to delete key: %v", w.Body.String())
	}

	// Delete bucket
	req = httptest.NewRequest("DELETE", "/testbucket", nil)
	req.Header.Set("API-KEY", apiKey)
	w = httptest.NewRecorder()
	userRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to delete bucket: %v", w.Body.String())
	}

	// Create a system bucket
	req = httptest.NewRequest("PUT", "/"+reservedBucket, nil)
	req.Header.Set("API-KEY", apiKey)
	w = httptest.NewRecorder()
	userRouter.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("System bucket allowed create: %v", w.Body.String())
	}

}

func TestMain(m *testing.M) {
	// Set the MASTER_API_KEY environment variable for testing

	// Run the tests
	code := m.Run()

	os.Exit(code)
}
