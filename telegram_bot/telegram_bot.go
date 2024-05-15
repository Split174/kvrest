package telegram_bot

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.etcd.io/bbolt"
)

const dataPath = "./data/"
const usersDbPath = dataPath + "users.db"

func StartBot() {
	userDB, err := bbolt.Open(usersDbPath, 0666, nil)
	if err != nil {
		log.Fatalf("Could not open users db: %s", err)
	}
	defer userDB.Close()

	userDB.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("users"))
		return err
	})

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
			case "create_kv":
				handleCreateKV(bot, userDB, update.Message)

			case "change_api_key":
				handleChangeApiKey(bot, userDB, update.Message)
			}
		}
	}
}

func handleCreateKV(bot *tgbotapi.BotAPI, db *bbolt.DB, msg *tgbotapi.Message) {
	userID := msg.From.ID
	var apiKey string

	db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("users"))
		apiKey = string(bucket.Get([]byte(fmt.Sprintf("%d", userID))))
		return nil
	})

	if apiKey != "" {
		response := "You already have an API key. Use /change_api_key to change it."
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, response))
		return
	}

	newApiKey, err := generateAPIKey()
	if err != nil {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Failed to generate API key"))
		return
	}

	db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("users"))
		return bucket.Put([]byte(fmt.Sprintf("%d", userID)), []byte(newApiKey))
	})

	os.OpenFile(filepath.Join(dataPath, newApiKey+".db"), os.O_RDONLY|os.O_CREATE, 0666)

	response := fmt.Sprintf("Your API key is: %s", newApiKey)
	bot.Send(tgbotapi.NewMessage(msg.Chat.ID, response))
}

func handleChangeApiKey(bot *tgbotapi.BotAPI, db *bbolt.DB, msg *tgbotapi.Message) {
	userID := msg.From.ID
	var oldApiKey string

	db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("users"))
		oldApiKey = string(bucket.Get([]byte(fmt.Sprintf("%d", userID))))
		return nil
	})

	if oldApiKey == "" {
		response := "You don't have an existing API key. Use /create_kv to create one."
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, response))
		return
	}

	newApiKey, err := generateAPIKey()
	if err != nil {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Failed to generate new API key"))
		return
	}

	db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("users"))
		return bucket.Put([]byte(fmt.Sprintf("%d", userID)), []byte(newApiKey))
	})

	oldDbFile := filepath.Join(dataPath, oldApiKey+".db")
	newDbFile := filepath.Join(dataPath, newApiKey+".db")
	err = os.Rename(oldDbFile, newDbFile)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Failed to rename database file"))
		return
	}

	response := fmt.Sprintf("Your new API key is: %s", newApiKey)
	bot.Send(tgbotapi.NewMessage(msg.Chat.ID, response))
}

func generateAPIKey() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
