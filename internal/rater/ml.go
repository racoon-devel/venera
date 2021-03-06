package rater

import (
	"fmt"
	"math"

	"github.com/BurntSushi/toml"
	"github.com/ccding/go-logging/logging"
	"github.com/racoon-devel/venera/internal/rater/classifier"
	"github.com/racoon-devel/venera/internal/types"
	"github.com/racoon-devel/venera/internal/utils"
)

type mlConfig struct {
	Threshold float32
}

type mlRater struct {
	c         *classifier.Classifier
	config    mlConfig
	log       *logging.Logger
	nextRater types.Rater
}

func (r *mlRater) Init(log *logging.Logger, settings *types.SearchSettings) {
	r.log = log

	path := utils.Configuration.Directories.Content + "/ml/"
	var err error
	r.c, err = classifier.NewClassifier(
		path+"retrained_graph.pb",
		path+"retrained_labels.txt",
		path+"frozen_inference_graph.pb",
	)

	if err != nil {
		panic(err)
	}

	path = fmt.Sprintf("%s/configurations/ml.conf", utils.Configuration.Directories.Content)
	_, err = toml.DecodeFile(path, &r.config)
	if err != nil {
		panic(err)
	}
}

func (r *mlRater) classify(data []byte) (float32, error) {
	img, err := r.c.PrepareImage(data)
	if err != nil {
		return 0, err
	}

	if img == nil {
		return 0, fmt.Errorf("person not recognizer")
	}

	rating, err := r.c.Classify(img)
	if err != nil {
		return 0, err
	}

	return rating, nil
}

func (r *mlRater) Rate(person *types.Person) int {

	var max float32

	for _, photo := range person.Photo {
		data, err := utils.HttpRequest(photo)
		if err != nil {
			r.log.Warnf("Retrieve image '%s' failed: %+v", photo, err)
			continue
		}

		rating, err := r.classify(data)
		if err != nil {
			r.log.Warnf("Classify image '%s' failed: %+v", photo, err)
		}

		if rating >= max {
			max = rating
		}
	}

	rating := 0
	if max >= r.config.Threshold {
		rating = int(math.Ceil(float64(max) * float64(100)))
	}

	return passNext(r.nextRater, person, rating)
}

func (r *mlRater) Next(nextRater types.Rater) types.Rater {
	r.nextRater = nextRater
	return nextRater
}

func (r *mlRater) Threshold(threshold int) int {
	return passThreshold(r.nextRater, threshold, int(math.Ceil(float64(r.config.Threshold)*100)))
}

func (r *mlRater) NeedPhotos() bool {
	return passNeedPhotos(r.nextRater, true)
}

func (r *mlRater) Close() {
	r.c.Close()
	propagateClose(r.nextRater)
}
