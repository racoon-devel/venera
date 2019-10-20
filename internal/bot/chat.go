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

func (self *chat) IncomingMessage(message string) tgbotapi.Chattable {
	args := strings.Split(message, " ")
	cmd, ok := commandSet[args[0]]
	if ok {
		msg, err := cmd.run(self.chat.ID, args[1:])
		if err != nil {
			return self.displayError(fmt.Errorf("Run command '%s' failed: %+v", args[0], err))
		}

		return msg
	}

	if len(args[0]) == 0 || args[0][0] == '/' {
		return self.displayError(fmt.Errorf("Command '%s' undefined", args[0]))
	}

	msg, _ := taskListHandler(self.chat.ID, args[1:])
	return msg
}

func (self *chat) displayError(err error) *tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(self.chat.ID, err.Error())
	log.Error(err)
	return &msg
}
