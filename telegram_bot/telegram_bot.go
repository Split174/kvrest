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
)

const dataPath = "./data/"

func StartBot() {
	/*userDB, err := bbolt.Open(usersDbPath, 0666, nil)
	if err != nil {
		log.Fatalf("Could not open users db: %s", err)
	}
	defer userDB.Close()

	userDB.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("users"))
		return err
	})*/

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
				handleCreateKV(bot, update.Message)

			case "change_api_key":
				handleChangeApiKey(bot, update.Message)
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
			bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "A KV alredy exist. Use /change_api_key"))
			return
		}
	}

	apiKey, err := generateAPIKey()
	if err != nil {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Failed to generate API key"))
		return
	}

	os.OpenFile(filepath.Join(dataPath, fmt.Sprintf("%d-%s.db", userID, apiKey)), os.O_RDONLY|os.O_CREATE, 0666)

	response := fmt.Sprintf("Your API key is: %s", fmt.Sprintf("%d-%s", userID, apiKey))
	bot.Send(tgbotapi.NewMessage(msg.Chat.ID, response))
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
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "KV does not exist. Use /create_kv"))
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

	response := fmt.Sprintf("Your new API key is: %s", fmt.Sprintf("%d-%s", userID, newApiKey))
	bot.Send(tgbotapi.NewMessage(msg.Chat.ID, response))
}

func generateAPIKey() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
