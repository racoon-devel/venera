package bot

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ccding/go-logging/logging"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"github.com/racoon-devel/venera/internal/utils"
)

type Channel chan *Message
type CommandHandler func(args []string, replyID string) (*Message, error)

type requestData struct {
	ctx             context.Context
	responseChannel chan string
}

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

const (
	requestTimeout = 10 * time.Minute
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

func Post(msg *Message) {
	bot.messageChan <- msg
}

func Request(ctx context.Context, text, image string) (string, error) {
	timeouted, _ := context.WithTimeout(ctx, requestTimeout)
	responseChannel := make(chan string)
	msg := newRequestMessage(timeouted, text, responseChannel)
	msg.Photo = image
	bot.messageChan <- msg

	select {
	case <-timeouted.Done():
		return "", timeouted.Err()
	case response := <-responseChannel:
		return response, nil
	}
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
	var replyID string

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
		replyID = update.CallbackQuery.ID

		// Некрасиво, но работает
		if strings.Index(userMessage, "/drop") == 0 {
			bot.api.DeleteMessage(tgbotapi.DeleteMessageConfig{ChatID: chatID, MessageID: update.CallbackQuery.Message.MessageID})
		}
	} else {
		return
	}

	if userName != bot.trustedUser {
		sendTextMessage(fmt.Sprintf("Not trusted user: %s", userName))
		return
	}

	bot.log.Debugf("[Message] %s: \"%s\" [%d, %d]", userName, userMessage, chatID, userID)

	if bot.trustedChat == nil {
		bot.trustedChat = newChat(telegramChat)
	}

	resp := bot.trustedChat.IncomingMessage(userMessage, replyID)
	if resp != nil {
		sendMessage(resp)
	}
}

func shutdown() {
	sendTextMessage("Shutdowning...")
	close(bot.messageChan)
}

func sendTextMessage(text string) {
	sendMessage(&Message{Content: text})
}

func sendMessage(message *Message) {
	if bot.trustedChat != nil {
		rawMessages := makeRawMessage(bot.trustedChat.ChatID(), message)
		for _, msg := range rawMessages {
			if _, err := bot.api.Send(msg); err != nil {
				bot.log.Errorf("Send message failed: %+v", err)
			}
		}

		if message.messageType == messageReply && message.replyID != "" {
			_, err := bot.api.AnswerCallbackQuery(tgbotapi.CallbackConfig{CallbackQueryID: message.replyID, Text: message.Content})
			if err != nil {
				bot.log.Errorf("Reply to message failed: %+v", err)
			}
		}

		if message.messageType == messageRequest {
			bot.trustedChat.Request(message.request)
		}
	}
}
