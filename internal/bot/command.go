package bot

import (
	"fmt"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	"racoondev.tk/gitea/racoon/venera/internal/dispatcher"
)

type commandHandler func(chatID int64, args []string) (*tgbotapi.MessageConfig, error)

type userCommand struct {
	run  commandHandler
	hint string
}

var (
	commandSet = map[string]userCommand{
		"/list":    {run: taskListHandler, hint: "Get Task List"},
		"/task":    {run: taskHandler, hint: "Get Task Info"},
		"/run":     {run: taskRunHandler, hint: "Run the task"},
		"/suspend": {run: taskSuspendHandler, hint: "Suspend the task"},
		"/stop":    {run: taskStopHandler, hint: "Stop the task"},
		"/delete":  {run: taskDeleteHandler, hint: "Remove the task"},
	}
)

func taskListHandler(chatID int64, args []string) (*tgbotapi.MessageConfig, error) {
	msg := tgbotapi.NewMessage(chatID, "Select task")
	buttons := make([]tgbotapi.InlineKeyboardButton, 0)
	tasks := dispatcher.Describe()
	for _, task := range tasks {
		button := tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("Task #%d [ %s ]", task.ID, task.Provider), fmt.Sprintf("/task %d", task.ID)),
		)

		buttons = append(buttons, button...)
	}

	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(buttons)
	return &msg, nil
}

func taskHandler(chatID int64, args []string) (*tgbotapi.MessageConfig, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("Missing taskID argument")
	}

	taskID := args[0]
	id, err := strconv.ParseUint(taskID, 10, 32)
	if err != nil {
		return nil, err
	}

	taskInfo, err := dispatcher.GetTaskInfo(uint(id))
	if err != nil {
		return nil, err
	}

	msg := tgbotapi.NewMessage(chatID, "Select action for task")
	buttons := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Stop", fmt.Sprintf("/stop %s", taskID)),
		tgbotapi.NewInlineKeyboardButtonData("Delete", fmt.Sprintf("/delete %s", taskID)),
	)

	if taskInfo.Mode == dispatcher.ModeIdle {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData("Run", fmt.Sprintf("/run %s", taskID)))
	} else {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData("Suspend", fmt.Sprintf("/suspend %s", taskID)))
	}

	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(buttons)

	return &msg, nil
}

func taskRunHandler(chatID int64, args []string) (*tgbotapi.MessageConfig, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("Missing taskID argument")
	}

	taskID, err := strconv.ParseUint(args[0], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("Invalid task ID: %s: %+v", args[0], err)
	}

	if err := dispatcher.RunTask(uint(taskID)); err != nil {
		return nil, err
	}

	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Task #%d started", taskID))
	return &msg, nil
}

func taskSuspendHandler(chatID int64, args []string) (*tgbotapi.MessageConfig, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("Missing taskID argument")
	}

	taskID, err := strconv.ParseUint(args[0], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("Invalid task ID: %s: %+v", args[0], err)
	}

	if err := dispatcher.SuspendTask(uint(taskID)); err != nil {
		return nil, err
	}

	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Task #%d suspended", taskID))
	return &msg, nil
}

func taskStopHandler(chatID int64, args []string) (*tgbotapi.MessageConfig, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("Missing taskID argument")
	}

	taskID, err := strconv.ParseUint(args[0], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("Invalid task ID: %s: %+v", args[0], err)
	}

	if err := dispatcher.StopTask(uint(taskID)); err != nil {
		return nil, err
	}

	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Task #%d stopped", taskID))
	return &msg, nil
}

func taskDeleteHandler(chatID int64, args []string) (*tgbotapi.MessageConfig, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("Missing taskID argument")
	}

	taskID, err := strconv.ParseUint(args[0], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("Invalid task ID: %s: %+v", args[0], err)
	}

	if err := dispatcher.DeleteTask(uint(taskID)); err != nil {
		return nil, err
	}

	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Task #%d deleted", taskID))
	return &msg, nil
}
