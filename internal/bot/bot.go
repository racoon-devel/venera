package bot

import (
	"fmt"
	"sync"

	"github.com/ccding/go-logging/logging"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"racoondev.tk/gitea/racoon/venera/internal/utils"
)

var log *logging.Logger
var done chan bool
var trustedChat *chat

func Init(logger *logging.Logger, APIToken string, trustedUser string, waitGroup *sync.WaitGroup) error {
	log = logger

	log.Infof("Start Telegram Bot { token = '%s' }", APIToken)

	bot, err := tgbotapi.NewBotAPIWithClient(APIToken, utils.GetHTTPClient())

	if err != nil {
		return fmt.Errorf("Create Telegram API instance failed: %+v", err)
	}

	done = make(chan bool)

	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		loop(bot, trustedUser)
	}()

	return nil

}

func loop(bot *tgbotapi.BotAPI, trustedUser string) {

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 0

	updates, err := bot.GetUpdatesChan(updateConfig)

	if err != nil {
		log.Errorf("Get bot updates failed: %+v", err)
		return
	}

	for {
		select {

		case update := <-updates:
			var chatID int64
			var userID int
			var userName string
			var userMessage string
			var telegramChat *tgbotapi.Chat

			if update.Message != nil {
				chatID = update.Message.Chat.ID
				telegramChat = update.Message.Chat
				userID = update.Message.From.ID
				userName = update.Message.From.UserName
				userMessage = update.Message.Text
			} else if update.CallbackQuery != nil {
				chatID = update.CallbackQuery.Message.Chat.ID
				userID = update.CallbackQuery.From.ID
				userName = update.CallbackQuery.From.UserName
				userMessage = update.CallbackQuery.Data
				telegramChat = update.CallbackQuery.Message.Chat

				bot.AnswerCallbackQuery(tgbotapi.CallbackConfig{CallbackQueryID: update.CallbackQuery.ID})
			} else {
				continue
			}

			if userName != trustedUser {
				msg := tgbotapi.NewMessage(chatID, "You are not trusted user")
				bot.Send(msg)
				continue
			}

			log.Debugf("[Message] %s: \"%s\" [%d, %d]", userName, userMessage, chatID, userID)

			if trustedChat == nil {
				trustedChat = newChat(telegramChat)
			}

			resp := trustedChat.IncomingMessage(userMessage)
			bot.Send(resp)

		case <-done:
			return
		}
	}
}
