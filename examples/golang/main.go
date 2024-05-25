package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type KVRestAPI struct {
	APIKey  string
	BaseURL string
	Client  *http.Client
}

type BucketListResponse struct {
	Buckets []string `json:"buckets"`
}

type KeyListResponse struct {
	Keys []string `json:"keys"`
}

func NewKVRestAPI(apiKey string) *KVRestAPI {
	return &KVRestAPI{
		APIKey:  apiKey,
		BaseURL: "https://kvrest.dev/api",
		Client:  &http.Client{},
	}
}

func (api *KVRestAPI) makeRequest(method, endpoint string, body interface{}) ([]byte, error) {
	url := fmt.Sprintf("%s%s", api.BaseURL, endpoint)
	var reqBody []byte

	if body != nil {
		var err error
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("API-KEY", api.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := api.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API request failed with status: %s, response: %s", resp.Status, string(respBody))
	}

	return respBody, nil
}

func (api *KVRestAPI) CreateBucket(bucketName string) error {
	_, err := api.makeRequest("PUT", "/"+bucketName, nil)
	return err
}

func (api *KVRestAPI) DeleteBucket(bucketName string) error {
	_, err := api.makeRequest("DELETE", "/"+bucketName, nil)
	return err
}

func (api *KVRestAPI) ListBuckets() ([]string, error) {
	respBody, err := api.makeRequest("POST", "/buckets", nil)
	if err != nil {
		return nil, err
	}

	var bucketResponse BucketListResponse
	err = json.Unmarshal(respBody, &bucketResponse)
	if err != nil {
		return nil, err
	}

	return bucketResponse.Buckets, nil
}

func (api *KVRestAPI) CreateOrUpdateKeyValue(bucketName, key string, value interface{}) error {
	_, err := api.makeRequest("PUT", fmt.Sprintf("/%s/%s", bucketName, key), value)
	return err
}

func (api *KVRestAPI) GetValue(bucketName, key string, target interface{}) error {
	respBody, err := api.makeRequest("GET", fmt.Sprintf("/%s/%s", bucketName, key), nil)
	if err != nil {
		return err
	}

	err = json.Unmarshal(respBody, target)
	return err
}

func (api *KVRestAPI) DeleteKeyValue(bucketName, key string) error {
	_, err := api.makeRequest("DELETE", fmt.Sprintf("/%s/%s", bucketName, key), nil)
	return err
}

func (api *KVRestAPI) ListKeys(bucketName string) ([]string, error) {
	respBody, err := api.makeRequest("GET", "/"+bucketName, nil)
	if err != nil {
		return nil, err
	}

	// Assuming the response is a JSON array of keys
	var keysResponse KeyListResponse
	err = json.Unmarshal(respBody, &keysResponse)
	if err != nil {
		return nil, err
	}

	return keysResponse.Keys, nil
}

func main() {
	apiKey := "CHANGE-ME" // Replace with your actual API key
	api := NewKVRestAPI(apiKey)

	bucketName := "my-test-bucket-go"
	key := "my-key"
	value := map[string]string{"message": "Hello from Go!"}

	// Create a bucket
	err := api.CreateBucket(bucketName)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Bucket '%s' created successfully.\n", bucketName)

	// Create a key-value pair
	err = api.CreateOrUpdateKeyValue(bucketName, key, value)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Key-value pair created/updated for key '%s'.\n", key)

	// Retrieve the value
	var retrievedValue map[string]string
	err = api.GetValue(bucketName, key, &retrievedValue)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Retrieved value for key '%s': %+v\n", key, retrievedValue)

	// List all buckets
	buckets, err := api.ListBuckets()
	if err != nil {
		panic(err)
	}
	fmt.Printf("All buckets: %+v\n", buckets)

	// List keys in the bucket
	keys, err := api.ListKeys(bucketName)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Keys in bucket '%s': %+v\n", bucketName, keys)

	// Delete the key-value pair
	err = api.DeleteKeyValue(bucketName, key)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Key-value pair with key '%s' deleted.\n", key)

	// Delete the bucket
	err = api.DeleteBucket(bucketName)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Bucket '%s' deleted.\n", bucketName)
}
