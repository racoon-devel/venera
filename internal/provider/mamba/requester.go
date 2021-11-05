package mamba

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/racoon-devel/venera/internal/utils"
)

const (
	CountryID = 3159
	RegionID  = 4312
)

type methodArgs map[string]string

type mambaResponse struct {
	Status  int
	Data    interface{}
	Message string
}

type mambaRequester struct {
	appID     uint
	secretKey string
}

func newMambaRequester(appID uint, secretKey string) *mambaRequester {
	return &mambaRequester{appID: appID, secretKey: secretKey}
}

func (m *mambaRequester) request(method string, args methodArgs, output interface{}) error {
	url := "http://api.aplatform.ru/?"

	args["app_id"] = strconv.Itoa(int(m.appID))
	args["secure"] = "1"
	args["method"] = method

	keys := make([]string, 0, len(args))
	for k := range args {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	hash := md5.New()
	for _, k := range keys {
		param := fmt.Sprintf("%s=%s", k, args[k])
		io.WriteString(hash, param)
		url += param + "&"
	}

	io.WriteString(hash, m.secretKey)

	url += "&sig=" + fmt.Sprintf("%x", hash.Sum(nil))

	data, err := utils.HttpRequest(url)
	if err != nil {
		return err
	}

	var response mambaResponse
	err = json.Unmarshal(data, &response)
	if err != nil {
		return err
	}

	if response.Status != 0 {
		return fmt.Errorf("Mamba API error: '%s'", response.Message)
	}

	err = mapstructure.Decode(response.Data, output)
	if err != nil {
		fmt.Println(string(data))
	}

	return err
}

type geoItem struct {
	ID   int
	Name string
}

func (m *mambaRequester) GetRegions(regions *[]geoItem) error {
	return m.request("geo.getRegions", methodArgs{"country_id": strconv.Itoa(CountryID)}, regions)
}

func (m *mambaRequester) GetCities(regionID int, cities *[]geoItem) error {
	return m.request("geo.getCities", methodArgs{"region_id": strconv.Itoa(regionID)}, cities)
}

func (m *mambaRequester) GetCityID(city string) (int, error) {
	cities := make([]geoItem, 0)
	if err := m.GetCities(RegionID, &cities); err != nil {
		return 0, err
	}

	target := strings.ToLower(city)
	for _, c := range cities {
		if strings.ToLower(c.Name) == target {
			return c.ID, nil
		}
	}

	return 0, fmt.Errorf("city '%s' not found", city)
}

type mambaUser struct {
	Info struct {
		Oid   int
		Name  string
		Link  string `mapstructure:"anketa_link"`
		Photo string `mapstructure:"medium_photo_url"`
		Age   int
		Sign  string
		Lang  string
	}

	PhotosCount int `mapstructure:"photos_count"`
	Flags       struct {
		IsVIP    int `mapstructure:"is_vip"`
		IsReal   int `mapstructure:"is_real"`
		IsLeader int `mapstructure:"is_leader"`
		Maketop  int
		IsOnline int `mapstructure:"is_online"`
	}
	About       string
	Familiarity struct {
		LookFor    string
		WaitingFor string
		Targets    []string
		Iam        struct {
			Stat string
		}
		Marital  string
		Children string
	}
	TypeDetail struct {
		Height       int
		Weight       int
		Circumstance string
		Constitution string
		Smoke        string
		Drink        string
		Home         string
		Language     []string
		Race         string
		Orientation  string
	} `mapstructure:"-"`
	Type      interface{}
	Interests []string
}

type searchResponse struct {
	Total int
	Limit int
	Users []interface{}
}

func (m *mambaRequester) Search(ageFrom uint, ageTo uint, city int, offset int) ([]mambaUser, error) {
	args := make(methodArgs)
	users := make([]mambaUser, 0)

	args["iam"] = "M"
	args["look_for"] = "F"
	args["age_from"] = strconv.Itoa(int(ageFrom))
	args["age_to"] = strconv.Itoa(int(ageTo))
	args["with_photo"] = "1"
	args["country_id"] = strconv.Itoa(CountryID)
	args["region_id"] = strconv.Itoa(RegionID)
	args["city_id"] = strconv.Itoa(city)
	args["blocks"] = "about, flags, familiarity, type, interests"
	args["offset"] = strconv.Itoa(offset)

	var response searchResponse

	if err := m.request("search.get", args, &response); err != nil {
		return nil, err
	}

	for i := range response.Users {
		user := mambaUser{}
		if err := mapstructure.Decode(response.Users[i], &user); err != nil {
			continue
		}

		mapstructure.Decode(user.Type, &user.TypeDetail)
		users = append(users, user)
	}

	return users, nil
}

func (m *mambaRequester) GetPhotos(oid int) ([]string, error) {
	type userPhotos struct {
		Photos []struct {
			Photo string `mapstructure:"huge_photo_url"`
		}
	}

	photos := userPhotos{}
	if err := m.request("photos.get", methodArgs{"oid": strconv.Itoa(oid)}, &photos); err != nil {
		return nil, err
	}

	result := make([]string, len(photos.Photos))
	for i, p := range photos.Photos {
		result[i] = p.Photo
	}

	albums, err := m.GetAlbums(oid)
	if err == nil {
		for _, album := range albums {
			nextPhotos := userPhotos{}
			if err := m.request("photos.get", methodArgs{"oid": strconv.Itoa(oid), "album_id": strconv.Itoa(album)}, &nextPhotos); err == nil {
				for _, p := range nextPhotos.Photos {
					result = append(result, p.Photo)
				}
			}
		}
	}

	return result, nil
}

func (m *mambaRequester) GetAlbums(oid int) ([]int, error) {
	type userAlbums struct {
		Albums []struct {
			AlbumID int `mapstructure:"album_id"`
		}
	}

	albums := userAlbums{}
	if err := m.request("photos.getAlbums", methodArgs{"oid": strconv.Itoa(oid)}, &albums); err != nil {
		return nil, err
	}

	result := make([]int, len(albums.Albums))
	for i, a := range albums.Albums {
		result[i] = a.AlbumID
	}

	return result, nil
}

func (m *mambaRequester) GetLastVisitTime(oids []int) ([]time.Time, error) {
	type userIsOnline struct {
		AnketaID int  `mapstructure:"anketa_id"`
		IsOnline uint `mapstructure:"is_online"`
	}

	list := ""
	for _, oid := range oids {
		list += strconv.Itoa(oid) + ","
	}

	userRec := make([]userIsOnline, 0)
	now := time.Now()

	list = strings.TrimSuffix(list, ",")
	if err := m.request("anketa.isOnline", methodArgs{"oids": list}, &userRec); err != nil {
		return nil, err
	}

	results := make([]time.Time, len(oids))
	for i, u := range userRec {
		if u.IsOnline > 2 {
			results[i] = time.Unix(int64(u.IsOnline), 0)
		} else {
			results[i] = now
		}
	}

	return results, nil
}
