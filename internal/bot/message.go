package bot

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"racoondev.tk/gitea/racoon/venera/internal/types"
)

const (
	messageSingle = iota
	messageMenu
	messageReply
	messageRequest
)

type Message struct {
	messageType int

	// messageReply section
	replyID string

	// messageRequest section
	request *requestData

	// Some content
	Content      string
	Photo        string
	PhotoCaption string
	Actions      []types.Action
}

func NewMenuMessage(content string, actions []types.Action) *Message {
	return &Message{messageType: messageMenu, Content: content, Actions: actions}
}

func NewReplyMessage(text string, replyID string) *Message {
	return &Message{messageType: messageReply, Content: text, replyID: replyID}
}

func newRequestMessage(ctx context.Context, text string, response chan string) *Message {
	return &Message{
		messageType: messageRequest,
		request:     &requestData{ctx: ctx, responseChannel: response},
		Content:     text,
	}
}

func makeRawMessage(chatID int64, message *Message) []tgbotapi.Chattable {

	switch message.messageType {
	case messageSingle:
		return makeSingleMessage(chatID, message)
	case messageMenu:
		return makeMenuMessage(chatID, message)
	case messageReply:
		if (message.replyID == "") {
			return makeSingleMessage(chatID, message)
		}
		fallthrough
	default:
		return make([]tgbotapi.Chattable, 0)
	}
}

func makeMenuMessage(chatID int64, message *Message) []tgbotapi.Chattable {
	msg := tgbotapi.NewMessage(chatID, message.Content)
	msg.ParseMode = "HTML"
	msg.Text = message.Content

	if message.Actions != nil && len(message.Actions) > 0 {
		buttons := make([]tgbotapi.KeyboardButton, 0)
		for _, action := range message.Actions {
			row := tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(action.Command),
			)

			buttons = append(buttons, row...)
		}

		keyboard := tgbotapi.NewReplyKeyboard(buttons)
		keyboard.OneTimeKeyboard = true
		msg.ReplyMarkup = keyboard

	}

	return []tgbotapi.Chattable{&msg}
}

func makeSingleMessage(chatID int64, message *Message) []tgbotapi.Chattable {
	messages := make([]tgbotapi.Chattable, 0)

	if message.Photo != "" {
		photo := tgbotapi.PhotoConfig{}
		photo.ChatID = chatID
		photo.FileID = message.Photo
		photo.UseExisting = true
		photo.Caption = message.PhotoCaption
		messages = append(messages, &photo)
	}

	msg := tgbotapi.NewMessage(chatID, message.Content)
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
	} else {
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(defaultCommand.Hint)),
		)
	}

	messages = append(messages, &msg)
	return messages
}
