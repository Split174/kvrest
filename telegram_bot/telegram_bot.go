package telegram_bot

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.etcd.io/bbolt"
)

const dataPath = "./data/"

func StartBot() {

	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "help":
				handleHelp(bot, update.Message)

			case "start":
				handleCreateKV(bot, update.Message)

			case "change_api_key":
				handleChangeApiKey(bot, update.Message)

			case "view_bucket_keys":
				handleViewBucketKeys(bot, update.Message)

			case "list_buckets":
				handleListBuckets(bot, update.Message)

			case "download_kv":
				handleDownloadKV(bot, update.Message)
			}
		}
	}
}

func handleCreateKV(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	userID := msg.From.ID

	// Construct the filename prefix based on the user's ID
	fileNamePrefix := fmt.Sprintf("%d-", userID)

	// Check if any file starting with the constructed prefix exists
	files, err := ioutil.ReadDir(dataPath)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		if strings.HasPrefix(file.Name(), fileNamePrefix) {
			apiKey := strings.Replace(file.Name(), ".db", "", 1)
			response := fmt.Sprintf("A KV already exists. Your API key: `%s`", apiKey)
			responseMsg := tgbotapi.NewMessage(msg.Chat.ID, response)
			responseMsg.ParseMode = "Markdown"
			bot.Send(responseMsg)
			return
		}
	}

	apiKey, err := generateAPIKey()
	if err != nil {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Failed to generate API key"))
		return
	}

	// Create the BoltDB file
	db, err := bbolt.Open(filepath.Join(dataPath, fmt.Sprintf("%d-%s.db", userID, apiKey)), 0666, nil)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Failed to create database file"))
		return
	}
	defer db.Close()

	response := fmt.Sprintf("Your API key is: `%s`", fmt.Sprintf("%d-%s", userID, apiKey))
	responseMsg := tgbotapi.NewMessage(msg.Chat.ID, response)
	responseMsg.ParseMode = "Markdown"
	bot.Send(responseMsg)
}

func handleChangeApiKey(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	userID := msg.From.ID
	fileNamePrefix := fmt.Sprintf("%d-", userID)

	var userDB string
	files, err := ioutil.ReadDir(dataPath)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		if strings.HasPrefix(file.Name(), fileNamePrefix) {
			userDB = file.Name()
			break
		}
	}

	if userDB == "" {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "KV does not exist. Use /start"))
		return
	}

	newApiKey, err := generateAPIKey()
	if err != nil {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Failed to generate new API key"))
		return
	}

	oldDbFile := filepath.Join(dataPath, userDB)
	newDbFile := filepath.Join(dataPath, fmt.Sprintf("%d-%s.db", userID, newApiKey))
	err = os.Rename(oldDbFile, newDbFile)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Failed to rename database file"))
		return
	}

	response := fmt.Sprintf("Your new API key is: `%s`", fmt.Sprintf("%d-%s", userID, newApiKey))
	responseMsg := tgbotapi.NewMessage(msg.Chat.ID, response)
	responseMsg.ParseMode = "Markdown"
	bot.Send(responseMsg)
}

func handleViewBucketKeys(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	userID := msg.From.ID
	fileNamePrefix := fmt.Sprintf("%d-", userID)

	var userDB string
	files, err := ioutil.ReadDir(dataPath)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		if strings.HasPrefix(file.Name(), fileNamePrefix) {
			userDB = file.Name()
			break
		}
	}

	if userDB == "" {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "You don't have a KV store. Use /start to create one."))
		return
	}

	// Get bucket name from the command arguments
	commandArgs := strings.Fields(msg.Text)
	if len(commandArgs) < 2 {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Please specify the bucket name using `/view_bucket BUCKET_NAME`"))
		return
	}
	bucketName := commandArgs[1]

	db, err := bbolt.Open(filepath.Join(dataPath, userDB), 0666, nil)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Failed to open database file"))
		return
	}
	defer db.Close()

	var bucketContent string
	err = db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			return fmt.Errorf("bucket '%s' not found", bucketName)
		}

		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			bucketContent += fmt.Sprintf("%s\n", k)
		}
		return nil
	})

	if err != nil {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Error: %s", err.Error())))
		return
	}

	if bucketContent == "" {
		bucketContent = fmt.Sprintf("Bucket '%s' is empty.", bucketName)
	}

	bot.Send(tgbotapi.NewMessage(msg.Chat.ID, bucketContent))
}

func handleListBuckets(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	userID := msg.From.ID
	fileNamePrefix := fmt.Sprintf("%d-", userID)

	var userDB string
	files, err := ioutil.ReadDir(dataPath)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		if strings.HasPrefix(file.Name(), fileNamePrefix) {
			userDB = file.Name()
			break
		}
	}

	if userDB == "" {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "You don't have a KV store. Use /start to create one."))
		return
	}

	db, err := bbolt.Open(filepath.Join(dataPath, userDB), 0666, nil)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Failed to open database file."))
		return
	}
	defer db.Close()

	var bucketList string
	err = db.View(func(tx *bbolt.Tx) error {
		return tx.ForEach(func(name []byte, _ *bbolt.Bucket) error {
			bucketList += fmt.Sprintf("- %s\n", name)
			return nil
		})
	})

	if err != nil {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Error listing buckets: %s", err.Error())))
		return
	}

	if bucketList == "" {
		bucketList = "You have no buckets."
	}

	bot.Send(tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Your buckets:\n%s", bucketList)))
}

func generateAPIKey() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func handleDownloadKV(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	userID := msg.From.ID
	fileNamePrefix := fmt.Sprintf("%d-", userID)

	var userDB string
	files, err := ioutil.ReadDir(dataPath)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		if strings.HasPrefix(file.Name(), fileNamePrefix) {
			userDB = file.Name()
			break
		}
	}

	if userDB == "" {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "You don't have a KV store. Use /start to create one."))
		return
	}

	dbPath := filepath.Join(dataPath, userDB)

	// Send the database file
	doc := tgbotapi.NewDocument(msg.Chat.ID, tgbotapi.FilePath(dbPath))
	_, err = bot.Send(doc)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("Error sending database file: %s", err.Error())))
		return
	}
}

func handleHelp(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	userDB := "YOUR-API-KEY"
	userID := msg.From.ID
	fileNamePrefix := fmt.Sprintf("%d-", userID)
	files, err := ioutil.ReadDir(dataPath)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		if strings.HasPrefix(file.Name(), fileNamePrefix) {
			userDB = strings.Replace(file.Name(), ".db", "", 1)
			break
		}
	}

	response := `<b>/help</b>
Displays the documentation for all available commands to the user.

<b>/start</b>
Creates a new key-value (KV) store for the user. It generates a unique API key and creates a new BoltDB file to store the user's data. The API key is then sent back to the user.

<b>/change_api_key</b>
Allows the user to change their existing API key. It generates a new API key, renames the BoltDB file with the new key, and sends the new API key to the user.

<b>/view_bucket_keys</b>
Allows the user to view the keys stored in a specific bucket within their KV store. The user needs to provide the name of the bucket they want to view.

<i>Usage:</i> <code>/view_bucket <b>BUCKET_NAME</b></code>

<b>/list_buckets</b>
Lists all the buckets that the user has created in their KV store.

<b>/download_kv</b>
Allows the user to download their entire KV store as a BoltDB file. The bot will send the file directly to the user.

<b>API Examples</b>

Create bucket
<code>curl -X PUT -H "API-KEY: ` + userDB + `" https://kvrest.dev/api/yourBucketName</code>

Delete bucket
<code>curl -X DELETE -H "API-KEY: ` + userDB + `" https://kvrest.dev/api/yourBucketName</code>

Create/Update Key-Value pair in bucket
<code>curl -X PUT -H "API-KEY: ` + userDB + `" -H "Content-Type: application/json" --data '{"key": "value"}' https://kvrest.dev/api/yourBucketName/yourKey</code>

Get value by key
<code>curl -X GET -H "API-KEY: ` + userDB + `" https://kvrest.dev/api/yourBucketName/yourKey</code>

Delete key-value pair in bucket
<code>curl -X DELETE -H "API-KEY: ` + userDB + `" https://kvrest.dev/api/yourBucketName/yourKey</code>`
	responseMsg := tgbotapi.NewMessage(msg.Chat.ID, response)
	responseMsg.ParseMode = "HTML"
	bot.Send(responseMsg)
}
