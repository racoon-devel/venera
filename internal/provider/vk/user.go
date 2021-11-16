package vk

import (
	"fmt"
	"github.com/SevereCloud/vksdk/v2/object"
	"github.com/racoon-devel/venera/internal/storage"
	"github.com/racoon-devel/venera/internal/types"
	"strconv"
	"time"
)

func bioAdd(p *types.Person, category string, content string) {
	if len(content) != 0 {
		p.Bio += fmt.Sprintf("%s: %s\n", category, content)
	}
}

func (session *searchSession) isRateableUser(user *object.UsersUser) bool {
	age := getAge(user.Bdate)
	if age != 0 && (age < session.state.Search.AgeFrom || age > session.state.Search.AgeTo) {
		return false
	}
	if user.City.ID != 0 && user.City.ID != session.state.CommonData.CityID {
		return false
	}
	return user.Sex == sexWoman &&
		(user.Relation == object.UserRelationSingle || user.Relation == object.UserRelationActivelySearching) &&
		time.Since(time.Unix(int64(user.LastSeen.Time), 0)) < expiredAccountThreshold &&
		storage.SearchPerson(session.provider.ID(), strconv.Itoa(user.ID)) == nil
}

func getAge(birthday string) uint {
	if len(birthday) != 0 {
		if date, err := time.Parse("2.1.2006", birthday); err == nil {
			return uint(time.Since(date).Hours() / (24. * 365))
		}
	}
	return 0
}

func convertPersonRecord(u *object.UsersUser) *types.Person {
	p := &types.Person{
		UserID: strconv.Itoa(u.ID),
		Name:   u.FirstName + " " + u.LastName,
		Link:   fmt.Sprintf("https://vk.com/id%d", u.ID),
	}

	if u.Personal.Alcohol == 1 || u.Personal.Alcohol == 2 {
		p.Alco = types.Negative
	} else if u.Personal.Alcohol == 3 || u.Personal.Alcohol == 4 {
		p.Alco = types.Neutral
	} else if u.Personal.Alcohol == 5 {
		p.Alco = types.Positive
	}

	if u.Personal.Smoking == 1 || u.Personal.Smoking == 2 {
		p.Smoke = types.Negative
	} else if u.Personal.Smoking == 3 || u.Personal.Smoking == 4 {
		p.Smoke = types.Neutral
	} else if u.Personal.Smoking == 5 {
		p.Smoke = types.Positive
	}

	if len(u.Personal.Religion) != 0 {
		p.Bio += u.Personal.Religion + "\n"
	}

	if len(u.Personal.InspiredBy) != 0 {
		p.Bio += u.Personal.InspiredBy + "\n"
	}

	switch u.Personal.PeopleMain {
	case 1:
		p.Bio += "ум и креативность\n"
	case 2:
		p.Bio += "доброта и честность\n"
	case 3:
		p.Bio += "красота и здоровье\n"
	case 4:
		p.Bio += "власть и богатство\n"
	case 5:
		p.Bio += "смелость и упорство\n"
	case 6:
		p.Bio += "юмор и жизнелюбие\n"
	}

	switch u.Personal.LifeMain {
	case 1:
		p.Bio += "семья и дети\n"
	case 2:
		p.Bio += "карьера и деньги\n"
	case 3:
		p.Bio += "развлечения и отдых\n"
	case 4:
		p.Bio += "наука и исследования\n"
	case 5:
		p.Bio += "совершенствование мира\n"
	case 6:
		p.Bio += "саморазвитие\n"
	case 7:
		p.Bio += "красота и искусство\n"
	case 8:
		p.Bio += "слава и влияние\n"
	}

	bioAdd(p, "О себе", u.About)
	bioAdd(p, "Интересы", u.Interests)
	bioAdd(p, "Книги", u.Books)
	bioAdd(p, "Музыка", u.Music)
	bioAdd(p, "Фильмы", u.Movies)
	bioAdd(p, "Сериалы", u.Tv)
	bioAdd(p, "Игры", u.Games)
	bioAdd(p, "Цитаты", u.Quotes)

	p.Photo = make([]string, 1)
	p.Photo[0] = u.PhotoMaxOrig

	p.Age = getAge(u.Bdate)

	return p
}
