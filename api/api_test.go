package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gorilla/mux"
)

var tempDir string

// setupDatabase creates a temporary directory for the database files during testing.
func setupDatabase() error {
	var err error
	tempDir, err = os.MkdirTemp("", "testdb")
	if err != nil {
		return err
	}
	SetDataPath(tempDir)
	SetUsersDbPath(filepath.Join(tempDir, "users.db"))
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

	adminRouter := mux.NewRouter()
	userRouter := mux.NewRouter()

	RegisterAdminRoutes(adminRouter)
	RegisterRoutes(userRouter)

	// Create KV store and receive API key
	body := map[string]string{"name": "testuser"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("PUT", "/create_kv", bytes.NewReader(bodyBytes))
	req.Header.Set("MASTER-API-KEY", "master-key")
	w := httptest.NewRecorder()
	adminRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to create KV store: %v", w.Body.String())
	}

	var response map[string]string
	json.NewDecoder(w.Body).Decode(&response)
	apiKey, exists := response["api_key"]
	if !exists || apiKey == "" {
		t.Fatalf("Expected an API key in response")
	}

	// Create a bucket
	req = httptest.NewRequest("PUT", "/testbucket", nil)
	req.Header.Set("API-Key", apiKey)
	w = httptest.NewRecorder()
	userRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to create bucket: %v", w.Body.String())
	}

	// Set a key-value pair
	value := map[string]interface{}{"name": "test"}
	valueBytes, _ := json.Marshal(value)
	req = httptest.NewRequest("PUT", "/testbucket/testkey", bytes.NewReader(valueBytes))
	req.Header.Set("API-Key", apiKey)
	w = httptest.NewRecorder()
	userRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to set key: %v", w.Body.String())
	}

	// Get the value
	req = httptest.NewRequest("GET", "/testbucket/testkey", nil)
	req.Header.Set("API-Key", apiKey)
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
	req.Header.Set("API-Key", apiKey)
	w = httptest.NewRecorder()
	userRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to delete key: %v", w.Body.String())
	}

	// Change API key
	req = httptest.NewRequest("PUT", "/change_api_key", bytes.NewReader(bodyBytes))
	req.Header.Set("MASTER-API-KEY", "master-key")
	w = httptest.NewRecorder()
	adminRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to change API key: %v", w.Body.String())
	}

	newAPIKeyResponse := map[string]string{}
	json.NewDecoder(w.Body).Decode(&newAPIKeyResponse)
	newAPIKey, exists := newAPIKeyResponse["api_key"]
	if !exists || newAPIKey == "" {
		t.Fatalf("Expected a new API key in response")
	}

	// Attempt to access with old API key — should fail
	req = httptest.NewRequest("GET", "/testbucket/testkey", nil)
	req.Header.Set("API-Key", apiKey)
	w = httptest.NewRecorder()
	userRouter.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		t.Fatalf("Old API key should not work after change")
	}

	// Access with new API key — should pass
	req = httptest.NewRequest("GET", "/testbucket/testkey", nil)
	req.Header.Set("API-Key", newAPIKey)
	w = httptest.NewRecorder()
	userRouter.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("Expected 404 after key deletion with new API key: %v", w.Body.String())
	}
}

func TestMain(m *testing.M) {
	// Set the MASTER_API_KEY environment variable for testing
	os.Setenv("MASTER_API_KEY", "master-key")

	// Run the tests
	code := m.Run()

	// Unset the MASTER_API_KEY environment variable
	os.Unsetenv("MASTER_API_KEY")

	os.Exit(code)
}
