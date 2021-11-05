package webui

import (
	"fmt"

	"github.com/racoon-devel/venera/internal/types"
	"github.com/racoon-devel/venera/internal/utils"
)

func DecorPerson(person *types.Person) string {
	content := fmt.Sprintf("<b>%s</b>\n\n<b>Age:</b> %d\n<b>Rating:</b> %d\n", person.Name,
		person.Age, person.Rating)

	if person.Job != "" {
		content += fmt.Sprintf("<b>Job:</b> %s\n", person.Job)
	}

	if person.School != "" {
		content += fmt.Sprintf("<b>School:</b> %s\n", person.Job)
	}

	content += "\n" + utils.Highlight(person.Bio, person.BioMatches, "<i>", "</i>")

	return content
}
