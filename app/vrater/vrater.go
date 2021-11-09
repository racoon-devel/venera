package main

import (
	"flag"
	"fmt"
	"github.com/ccding/go-logging/logging"
	"github.com/racoon-devel/venera/internal/rater"
	"github.com/racoon-devel/venera/internal/storage"
	"github.com/racoon-devel/venera/internal/types"
	"github.com/racoon-devel/venera/internal/utils"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
)

func main() {
	fmt.Println("Venera Rater")

	runtime.GOMAXPROCS(runtime.NumCPU())

	logger, _ := logging.WriterLogger("vrater", logging.DEBUG, "%12s [%s][%7s:%3d] %s\n time,levelname,filename,lineno,message",
		"15:04:05.999", os.Stdout, true)

	defer logger.Destroy()

	configPath := flag.String("config", "/etc/venera/venera.conf",
		"set server configuration file")

	raterID := flag.String("rater", "default", "rate engine, can be default|ml|default+ml")
	profile := flag.String("profile", "", "set configuration profile")
	likes := flag.String("likes", "", "set likes test file")
	dislikes := flag.String("dislikes", "", "set dislikes test file")

	flag.Parse()

	logger.Infof("configuration file '%s' used", *configPath)

	if err := utils.Configuration.Load(*configPath); err != nil {
		logger.Critical(err)
		os.Exit(1)
	}

	logger.Debug(utils.Configuration)

	searchSettings, err := loadSearchSettings(*likes, *dislikes)
	if err != nil {
		logger.Criticalf("Load search settings failed: %+v", err)
		os.Exit(1)
	}

	r := rater.NewRater(*raterID, *profile, logger, searchSettings)

	if err := storage.Connect(utils.Configuration.GetConnectionString()); err != nil {
		logger.Critical(err)
		os.Exit(1)
	}

	const selectLimit = 100
	var offset uint = 0

	for {
		persons, total, err := storage.LoadPersons(0, false, selectLimit, 0, false, 0)
		if err != nil {
			panic(err)
		}
		for _, p := range persons {
			rating := r.Rate(&p.Person)
			logger.Infof("Person '%s' rating change %d => %d", p.Person.Name, p.Rating, rating)
			//storage.UpdateRating(p.ID, rating)
			p.Rating = rating
			storage.UpdatePerson(&p)

		}
		offset += uint(len(persons))
		if offset >= total {
			break
		}
	}
}

func loadSearchSettings(likes, dislikes string) (*types.SearchSettings, error) {
	settings := types.SearchSettings{}

	data, err := ioutil.ReadFile(likes)
	if err != nil {
		return nil, err
	}
	settings.Likes = strings.Split(string(data), ",")

	data, err = ioutil.ReadFile(dislikes)
	if err != nil {
		return nil, err
	}
	settings.Dislikes = strings.Split(string(data), ",")

	utils.TrimArray(settings.Likes)
	utils.TrimArray(settings.Dislikes)

	return &settings, nil
}
