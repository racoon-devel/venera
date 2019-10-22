package types

type Message struct {
	Content      string
	Photo        string
	PhotoCaption string
	Actions      []Action
}

type BotChannel chan *Message
