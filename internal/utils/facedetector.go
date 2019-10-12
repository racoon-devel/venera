package utils

import (
	"bytes"
	"image"
	"io/ioutil"

	pigo "github.com/esimov/pigo/core"
)

type FaceDetector struct {
	classifier *pigo.Pigo
}

func NewFaceDetector(cascadePath string) (*FaceDetector, error) {
	cascadeFile, err := ioutil.ReadFile(cascadePath)
	if err != nil {
		return nil, err
	}

	detector := FaceDetector{}

	pigo := pigo.NewPigo()
	detector.classifier, err = pigo.Unpack(cascadeFile)
	if err != nil {
		return nil, err
	}

	return &detector, nil
}

func loadImage(photo []byte) (*image.NRGBA, error) {
	reader := bytes.NewReader(photo)
	source, _, err := image.Decode(reader)
	if err != nil {
		return nil, err
	}

	return pigo.ImgToNRGBA(source), nil
}

func (self *FaceDetector) IsFacePresent(photo []byte) (bool, error) {

	source, err := loadImage(photo)
	if err != nil {
		return false, err
	}

	pixels := pigo.RgbToGrayscale(source)
	cols, rows := source.Bounds().Max.X, source.Bounds().Max.Y

	cParams := pigo.CascadeParams{
		MinSize:     20,
		MaxSize:     1000,
		ShiftFactor: 0.1,
		ScaleFactor: 1.1,

		ImageParams: pigo.ImageParams{
			Pixels: pixels,
			Rows:   rows,
			Cols:   cols,
			Dim:    cols,
		},
	}

	dets := self.classifier.RunCascade(cParams, 0.0)
	return len(dets) != 0, nil
}
