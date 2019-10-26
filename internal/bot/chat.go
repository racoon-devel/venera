package bot

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type chat struct {
	chat *tgbotapi.Chat
}

func newChat(telegramChat *tgbotapi.Chat) *chat {
	return &chat{chat: telegramChat}
}

func (self *chat) IncomingMessage(message string) *Message {
	cmdMutex.Lock()
	defer cmdMutex.Unlock()

	args := strings.Split(message, " ")
	if args[0][0] == '/' {
		cmd, ok := commandSet[args[0][1:]]
		if ok {
			if len(args)-1 < int(cmd.Args) {
				return displayError(fmt.Errorf("Run command '%s' failed: not enough actual parameters", args[0]))
			}
			msg, err := cmd.Run(args[1:])
			if err != nil {
				return displayError(fmt.Errorf("Run command '%s' failed: %+v", args[0], err))
			}

			return msg
		} else {
			return displayError(fmt.Errorf("Command '%s' undefined", args[0]))
		}
	}

	if defaultCommand != nil {
		msg, _ := defaultCommand.Run([]string{})
		return msg
	}

	return nil
}

func (self *chat) ChatID() int64 {
	return self.chat.ID
}

func displayError(err error) *Message {
	msg := Message{Content: err.Error()}
	bot.log.Error(err)
	return &msg
}
