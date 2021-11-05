package dispatcher

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/racoon-devel/venera/internal/webui"

	"github.com/racoon-devel/venera/internal/bot"
	"github.com/racoon-devel/venera/internal/storage"
	"github.com/racoon-devel/venera/internal/types"
)

func registerBotCommands() {
	bot.RegisterCommand("list", 0, "Get all tasks", taskListHandler)
	bot.RegisterCommand("task", 1, "Task control", taskHandler)
	bot.RegisterCommand("run", 1, "Run the task", taskRunHandler)
	bot.RegisterCommand("suspend", 1, "Suspend the task", taskSuspendHandler)
	bot.RegisterCommand("stop", 1, "Stop the task", taskStopHandler)
	bot.RegisterCommand("delete", 1, "Delete the task", taskDeleteHandler)
	bot.RegisterCommand("favour", 1, "Add person to favourites", personFavourHandler)
	bot.RegisterCommand("drop", 1, "Remove person record", personDropHandler)
	bot.RegisterCommand("action", 2, "Provider specific action", actionHandler)

	bot.SetDefaultCommand("list")
}

func taskListHandler(args []string, replyID string) (*bot.Message, error) {
	tasks := Describe()
	if len(tasks) == 0 {
		return &bot.Message{Content: "No tasks"}, nil
	}

	actions := make([]types.Action, 0)
	content := "Select task:\n"

	for _, task := range tasks {
		action := types.Action{
			Title:   fmt.Sprintf("Task #%d [ %s ]", task.ID, task.Provider),
			Command: fmt.Sprintf("/task %d", task.ID),
		}

		actions = append(actions, action)
		content += action.Title + " " + webui.StatusToHumanReadable(task.Status) + "\n"
	}

	return bot.NewMenuMessage(content, actions), nil
}

func taskHandler(args []string, replyID string) (*bot.Message, error) {
	taskID := args[0]
	id, err := strconv.ParseUint(taskID, 10, 32)
	if err != nil {
		return nil, err
	}

	taskInfo, err := GetTaskInfo(uint(id))
	if err != nil {
		return nil, err
	}

	actions := []types.Action{
		types.Action{Title: "Stop", Command: fmt.Sprintf("/stop %s", taskID)},
		types.Action{Title: "Delete", Command: fmt.Sprintf("/delete %s", taskID)},
	}

	if taskInfo.Mode == ModeIdle {
		actions = append(actions, types.Action{Title: "Run", Command: fmt.Sprintf("/run %s", taskID)})
	} else {
		actions = append(actions, types.Action{Title: "Suspend", Command: fmt.Sprintf("/suspend %s", taskID)})
	}

	return bot.NewMenuMessage("Select action for task", actions), nil
}

func taskControlHandler(arg string, handler func(taskID uint) error, msg string) (*bot.Message, error) {
	taskID, err := strconv.ParseUint(arg, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("Invalid task ID: %s: %+v", arg, err)
	}

	if err := handler(uint(taskID)); err != nil {
		return nil, err
	}

	return &bot.Message{Content: fmt.Sprintf("Task #%d %s", taskID, msg)}, nil
}

func taskRunHandler(args []string, replyID string) (*bot.Message, error) {
	return taskControlHandler(args[0], RunTask, "started")
}

func taskSuspendHandler(args []string, replyID string) (*bot.Message, error) {
	return taskControlHandler(args[0], SuspendTask, "suspended")
}

func taskStopHandler(args []string, replyID string) (*bot.Message, error) {
	return taskControlHandler(args[0], StopTask, "stopped")
}

func taskDeleteHandler(args []string, replyID string) (*bot.Message, error) {
	return taskControlHandler(args[0], DeleteTask, "deleted")
}

func personFavourHandler(args []string, replyID string) (*bot.Message, error) {
	personID, err := strconv.ParseUint(args[0], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("Invalid person ID: %s: %+v", args[0], err)
	}

	storage.Favourite(uint(personID))
	return bot.NewReplyMessage("Done", replyID), nil
}

func personDropHandler(args []string, replyID string) (*bot.Message, error) {
	personID, err := strconv.ParseUint(args[0], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("Invalid person ID: %s: %+v", args[0], err)
	}

	storage.DeletePerson(uint(personID))
	return bot.NewReplyMessage("Done", replyID), nil
}

func actionHandler(args []string, replyID string) (*bot.Message, error) {
	taskID, err := strconv.ParseUint(args[0], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("Invalid task ID: %s: %+v", args[0], err)
	}
	var params url.Values
	params = make(map[string][]string)

	for i := 2; i < len(args); i += 2 {
		param := make([]string, 1)
		param[0] = args[i+1]
		params[args[i]] = param
	}

	err = Action(uint(taskID), args[1], params)
	if err != nil {
		return nil, err
	}

	return bot.NewReplyMessage("Done", replyID), nil
}
