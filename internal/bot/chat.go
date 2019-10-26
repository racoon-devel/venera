package bot

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	stateMessaging = iota
	stateRequesting
)

type chat struct {
	chat *tgbotapi.Chat
	state int
	request *requestData
}

func newChat(telegramChat *tgbotapi.Chat) *chat {
	return &chat{chat: telegramChat}
}

func (self *chat) IncomingMessage(message string, replyID string) *Message {
	cmdMutex.Lock()
	defer cmdMutex.Unlock()

	if self.state == stateRequesting {
		self.request.responseChannel <- message
		self.request = nil
		self.state = stateMessaging
		return nil
	}

	args := strings.Split(message, " ")
	if args[0][0] == '/' {
		cmd, ok := commandSet[args[0][1:]]
		if ok {
			if len(args)-1 < int(cmd.Args) {
				return displayError(fmt.Errorf("Run command '%s' failed: not enough actual parameters", args[0]), replyID)
			}
			msg, err := cmd.Run(args[1:], replyID)
			if err != nil {
				return displayError(fmt.Errorf("Run command '%s' failed: %+v", args[0], err), replyID)
			}

			return msg
		} else {
			return displayError(fmt.Errorf("Command '%s' undefined", args[0]), replyID)
		}
	}

	if defaultCommand != nil {
		msg, _ := defaultCommand.Run([]string{}, "")
		return msg
	}

	return nil
}

func (self *chat) ChatID() int64 {
	return self.chat.ID
}

func(self *chat) Request(request *requestData) {
	if self.state == stateRequesting {
		// хуй знает, что делать
	}

	self.state = stateRequesting
	self.request = request
}

func displayError(err error, replyID string) *Message {

	bot.log.Error(err)
	if replyID == "" {
		return &Message{Content: err.Error()}
	}

	return NewReplyMessage(err.Error(), replyID)
}
