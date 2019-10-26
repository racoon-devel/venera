package bot

import "racoondev.tk/gitea/racoon/venera/internal/types"

type Message struct {
	Content      string
	Photo        string
	PhotoCaption string
	Actions      []types.Action
}

type Channel chan *Message

type CommandHandler func(args []string) (*Message, error)
