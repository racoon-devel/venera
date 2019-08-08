package dispatcher

import "github.com/ccding/go-logging/logging"

// SearchSession - search session of some provider
type SearchSession interface {
	SaveState() interface{}
	LoadState() interface{}
}

var dispatcher struct {
	log *logging.Logger
}

func Init(log *logging.Logger) {
	dispatcher.log = log
}

func NewTask(session SearchSession) {

}
