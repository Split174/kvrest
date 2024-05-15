# Key-Value Store API with Telegram Bot Integration

This project provides a key-value store API using Go and bbolt for database storage. Users can create and modify their API keys via a Telegram bot. Each user has their own separate database file.


## Usage

```
git clone https://github.com/Split174/kvest.git
cd kvest
mkdir -p data
export BOT_TOKEN=<YOUR_TOKEN>  && go build && ./kvest
```

## Telegram Bot Commands

1. Create Key-Value Store:
- Command: /create_kv `Generates a new API key and creates a corresponding database file for the user. Can only be used once per user.`
2. Change API Key:
- Command: /change_api_key `Changes the API key for the user and renames the database file to match the new API key.`


## API Endpoints
1. Create Bucket:
-   Endpoint: PUT /api/{bucketName}
-   Description: Create a new bucket.
-   Headers:
    - API-Key: {user's API key}

2. Set Key-Value:
- Endpoint: PUT /api/{bucketName}/{key}
- Description: Set a key-value pair in the specified bucket.
- Headers:
    - API-Key: {user's API key}
- Body: JSON object containing the value.

3. Get Value:
- Endpoint: GET /api/{bucketName}/{key}
- Description: Get the value of a key in the specified bucket.
- Headers:
    - API-Key: {user's API key}
4. Delete Key:
- Endpoint: DELETE /api/{bucketName}/{key}
- Description: Delete a key-value pair in the specified bucket.
- Headers:
    - API-Key: {user's API key}

## Project Structure

./project
├── go.mod
├── go.sum
├── main.go
├── telegram_bot
│   └── telegram_bot.go
├── telegram_bot.go
├── api
    └── api.go

## Requirements

- Go 1.16 or later
- A valid Telegram Bot Token

## Dependencies

This project uses the following Go packages:
- [bbolt](https://github.com/etcd-io/bbolt)
- [gorilla/mux](https://github.com/gorilla/mux)
- [go-telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api)

