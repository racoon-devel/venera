package tinder

import "math/rand"

const (
	latitudeRadius  float32 = 0.135568
	longitudeRadius float32 = 0.20243
	maxRadius       int     = 10
	maxShiftStep    float32 = 0.019
)

type geoposition struct {
	Latitude  float32
	Longitude float32
}

func randomPosition(center geoposition) geoposition {
	return geoposition{
		center.Latitude - latitudeRadius + rand.Float32()*2*latitudeRadius,
		center.Longitude - longitudeRadius + rand.Float32()*2*longitudeRadius,
	}
}

func randRadius() int {
	return 1 + rand.Intn(maxRadius)
}

func shiftPosition(center geoposition, position *geoposition) {
	latSign := 1
	lonSign := 1

	if rand.Intn(2) == 0 {
		latSign = -1
	}

	if rand.Intn(2) == 0 {
		lonSign = -1
	}

	position.Latitude = position.Latitude + float32(latSign)*(rand.Float32()*maxShiftStep)
	position.Longitude = position.Longitude + float32(lonSign)*(rand.Float32()*maxShiftStep)

	if position.Latitude > center.Latitude+latitudeRadius || position.Latitude < center.Latitude-latitudeRadius {
		position.Latitude = center.Latitude
	}

	if position.Longitude > center.Longitude+longitudeRadius || position.Longitude < center.Longitude-longitudeRadius {
		position.Longitude = center.Longitude
	}
}
