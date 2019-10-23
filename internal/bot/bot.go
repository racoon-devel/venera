package bot

import (
	"context"
	"fmt"
	"sync"

	"github.com/ccding/go-logging/logging"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"racoondev.tk/gitea/racoon/venera/internal/utils"
)

var bot struct {
	ctx         context.Context
	log         *logging.Logger
	wg          *sync.WaitGroup
	api         *tgbotapi.BotAPI
	trustedUser string
	trustedChat *chat
	messageChan Channel
}

var (
	cmdMutex       sync.Mutex
	commandSet     = make(map[string]botCommand, 0)
	defaultCommand *botCommand
)

func Initialize(ctx context.Context, logger *logging.Logger, wg *sync.WaitGroup,
	APIToken string, trustedUser string) error {

	bot.log = logger
	bot.ctx = ctx
	bot.wg = wg
	bot.trustedUser = trustedUser

	bot.log.Infof("Start Telegram Bot { token = '%s' }", APIToken)

	var err error
	bot.api, err = tgbotapi.NewBotAPIWithClient(APIToken, utils.GetHTTPClient())

	if err != nil {
		return fmt.Errorf("Create Telegram API instance failed: %+v", err)
	}

	bot.messageChan = make(Channel, 1000)

	wg.Add(1)
	go func() {
		defer bot.wg.Done()

		botLoop()
	}()

	return nil

}

func botLoop() {

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
			handleUpdates(&update)

		case <-bot.ctx.Done():
			shutdown()
			return

		case message := <-bot.messageChan:
			sendMessage(message)
		}
	}
}

func handleUpdates(update *tgbotapi.Update) {
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
		return
	}

	if userName != bot.trustedUser {
		sendTextMessage("You are not trusted user")
		return
	}

	bot.log.Debugf("[Message] %s: \"%s\" [%d, %d]", userName, userMessage, chatID, userID)

	if bot.trustedChat == nil {
		bot.trustedChat = newChat(telegramChat)
	}

	resp := bot.trustedChat.IncomingMessage(userMessage)
	if resp != nil {
		sendMessage(resp)
	}
}

func shutdown() {
	sendTextMessage("Shutdowning...")
	close(bot.messageChan)
}

func sendTextMessage(text string) {
	if bot.trustedChat != nil {
		msg := tgbotapi.NewMessage(bot.trustedChat.ChatID(), text)
		msg.ParseMode = "HTML"
		bot.api.Send(msg)
	}
}

func sendMessage(message *Message) {
	if bot.trustedChat != nil {
		if message.Photo != "" {
			photo := tgbotapi.PhotoConfig{}
			photo.ChatID = bot.trustedChat.ChatID()
			photo.FileID = message.Photo
			photo.UseExisting = true
			photo.Caption = message.PhotoCaption
			bot.api.Send(photo)
		}

		msg := tgbotapi.NewMessage(bot.trustedChat.ChatID(), message.Content)
		msg.ParseMode = "HTML"
		if message.Actions != nil && len(message.Actions) > 0 {
			buttons := make([]tgbotapi.InlineKeyboardButton, 0)
			for _, action := range message.Actions {
				row := tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(action.Title, action.Command),
				)

				buttons = append(buttons, row...)
			}

			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(buttons)
		}
		bot.api.Send(msg)
	}
}

func Post(msg *Message) {
	bot.messageChan <- msg
}
