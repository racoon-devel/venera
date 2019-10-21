package bot

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ccding/go-logging/logging"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"racoondev.tk/gitea/racoon/venera/internal/dispatcher"
	"racoondev.tk/gitea/racoon/venera/internal/utils"
)

var bot struct {
	ctx         context.Context
	log         *logging.Logger
	wg          *sync.WaitGroup
	api         *tgbotapi.BotAPI
	timer       *time.Ticker
	trustedChat *chat
	messageChan chan *botMessage
}

const (
	statInterval = time.Minute
)

type botMessage struct {
	Content string
}

func Init(ctx context.Context, logger *logging.Logger, wg *sync.WaitGroup, APIToken string, trustedUser string) error {
	bot.log = logger
	bot.ctx = ctx
	bot.wg = wg

	bot.log.Infof("Start Telegram Bot { token = '%s' }", APIToken)

	var err error
	bot.api, err = tgbotapi.NewBotAPIWithClient(APIToken, utils.GetHTTPClient())

	if err != nil {
		return fmt.Errorf("Create Telegram API instance failed: %+v", err)
	}

	bot.messageChan = make(chan *botMessage, 1000)

	bot.timer = time.NewTicker(statInterval)

	wg.Add(1)
	go loop(trustedUser)

	return nil

}

func loop(trustedUser string) {
	defer bot.wg.Done()
	defer bot.timer.Stop()

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 0

	updates, err := bot.api.GetUpdatesChan(updateConfig)

	if err != nil {
		bot.log.Errorf("Get bot updates failed: %+v", err)
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

				bot.api.AnswerCallbackQuery(tgbotapi.CallbackConfig{CallbackQueryID: update.CallbackQuery.ID})
			} else {
				continue
			}

			if userName != trustedUser {
				msg := tgbotapi.NewMessage(chatID, "You are not trusted user")
				bot.api.Send(msg)
				continue
			}

			bot.log.Debugf("[Message] %s: \"%s\" [%d, %d]", userName, userMessage, chatID, userID)

			if bot.trustedChat == nil {
				bot.trustedChat = newChat(telegramChat)
			}

			resp := bot.trustedChat.IncomingMessage(userMessage)
			bot.api.Send(resp)

		case <-bot.ctx.Done():
			sendStat()
			sendTextMessage("Shutdowning...")
			close(bot.messageChan)
			return

		case message := <-bot.messageChan:
			sendTextMessage(message.Content)

		case <-bot.timer.C:
			sendStat()
		}
	}
}

func sendTextMessage(text string) {
	if bot.trustedChat != nil {
		msg := tgbotapi.NewMessage(bot.trustedChat.ChatID(), text)
		msg.ParseMode = "HTML"
		bot.api.Send(msg)
	}
}

func sendStat() {
	stat := dispatcher.CollectStat()
	for _, text := range stat {
		bot.log.Info(text)
		sendTextMessage(text)
	}
}

func Post(message string) {
	msg := &botMessage{Content: message}
	bot.messageChan <- msg
}
