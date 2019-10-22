package interactive

import (
	"fmt"

	"racoondev.tk/gitea/racoon/venera/internal/utils"

	"github.com/ccding/go-logging/logging"
	"racoondev.tk/gitea/racoon/venera/internal/storage"
	"racoondev.tk/gitea/racoon/venera/internal/types"
)

var (
	log *logging.Logger
	db  *storage.Storage
	bot types.BotChannel
)

// Этот пакет предоставляет расширенное API для сессий поиска (архитектура - полный зашквар)

func Initialize(logger *logging.Logger, database *storage.Storage, botChannel types.BotChannel) {
	log = logger
	db = database
	bot = botChannel
}

func PostResult(provider types.Provider, person *types.Person) {
	rawID := provider.ID() + "." + person.UserID
	log.Debugf("Post person '%s' [ %s ]", person.Name, rawID)

	record := db.SearchPerson(rawID)
	if record == nil {
		log.Errorf("Person '%s' [ %s ] not found in database", person.Name, rawID)
		return
	}

	actions := provider.GetResultActions(record)

	dropAction := types.Action{Title: "Drop", Command: fmt.Sprintf("/drop %d", record.ID)}
	actions = append(actions, dropAction)

	if !record.Favourite {
		favouriteAction := types.Action{Title: "Add to Favourites", Command: fmt.Sprintf("/favour %d", record.ID)}
		actions = append(actions, favouriteAction)
	}

	content := fmt.Sprintf("<b>%s</b>\n\n<b>Age:</b> %d\n<b>Rating:</b> %d\n", person.Name,
		person.Age, person.Rating)

	if person.Job != "" {
		content += fmt.Sprintf("<b>Job:</b> %s\n", person.Job)
	}

	if person.School != "" {
		content += fmt.Sprintf("<b>School:</b> %s\n", person.Job)
	}

	content += "\n" + utils.Highlight(person.Bio, person.BioMatches, "<i>", "</i>")

	bot <- &types.Message{Content: content, Actions: actions, Photo: person.Photo[0],
		PhotoCaption: person.Name}
}
