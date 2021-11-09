package tinder

import (
	"fmt"
	"strings"
	"time"

	tindergo "github.com/racoon-devel/TinderGo"
	"github.com/racoon-devel/venera/internal/types"
)

func parseField(field map[string]interface{}) string {
	var result string
	for _, value := range field {
		mapValue, ok := value.(map[string]interface{})
		if ok {
			result += parseField(mapValue)
			continue
		}

		strValue, ok := value.(string)
		if ok {
			result += strValue + ","
		}
	}

	return result
}

func parseFieldList(field []interface{}) string {
	var info string
	for _, item := range field {
		m, ok := item.(map[string]interface{})
		if !ok {
			info += fmt.Sprint(item) + ","
			continue
		}

		info += parseField(m)
	}

	if len(info) != 0 {
		info = strings.TrimSuffix(info, info[len(info)-1:])
	}

	return info
}

func convertPersonRecord(record *tindergo.RecsCoreUser) types.Person {
	person := types.Person{UserID: record.ID, Name: record.Name, Bio: record.Bio}
	diff := time.Now().Sub(record.BirthDate)
	person.Age = uint(diff.Seconds() / 31207680)
	person.Job = parseFieldList(record.Jobs)
	person.School = parseFieldList(record.Schools)
	//person.VisitTime = record.PingTime

	person.Photo = make([]string, 0)
	for _, photo := range record.Photos {
		person.Photo = append(person.Photo, photo.URL)
	}

	return person
}

func convertMatch(match *tindergo.Match) types.Person {
	person := types.Person{UserID: match.Person.ID, Name: match.Person.Name, Bio: match.Person.Bio}
	diff := time.Now().Sub(match.Person.BirthDate)
	person.Age = uint(diff.Seconds() / 31207680)

	person.Photo = make([]string, 0)
	for _, photo := range match.Person.Photos {
		person.Photo = append(person.Photo, photo.URL)
	}

	return person
}
