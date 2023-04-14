package gps

import (
	"fmt"
	"math"
	"strconv"

	gpx "github.com/sudhanshuraheja/go-garmin-gpx"
)

type Location struct {
	Longitude float64
	Latitude  float64
}

type Route struct {
	Points []Location
}

func NewRoute() *Route {
	return &Route{
		// Create route with capacity to hold 5 mins
		// (if location updates once per 3 seconds)
		Points: make([]Location, 0, 5*(60/3)),
	}
}

func (r *Route) AddPoint(location Location) {
	// TODO: this can grow infinetely -> handle this
	r.Points = append(r.Points, location)
}

func (r *Route) Direction() float64 {
	// If you want to calculate direction as last point - point before that
	// set this to 1. Currently we do last point - 2 points before that in
	// order to have smoother map rotation
	lastNPointToCompare := 2

	if len(r.Points) <= lastNPointToCompare+1 {
		return 0.0
	}
	prevLoc := r.Points[len(r.Points)-lastNPointToCompare-1]
	currentLoc := r.Points[len(r.Points)-1]
	return Angle(
		prevLoc.Longitude,
		prevLoc.Latitude,
		currentLoc.Longitude,
		currentLoc.Latitude,
	)
}

type number interface {
	int | float64
}

// Angle Find the Angle between 2 points (considering top-left as 0, 0)
// Taken from here: https://stackoverflow.com/a/27481611
func Angle[T number](x1, y1, x2, y2 T) float64 {
	// NOTE: Remember that most math has the Y axis as positive above the X.
	// However, for screens we have Y as positive below. For this reason,
	// the Y values are inverted to get the expected results.
	deltaY := float64(y1 - y2)
	deltaX := float64(x2 - x1)

	resultRadians := math.Atan2(deltaY, deltaX)
	resultDegrees := resultRadians * (180 / math.Pi)

	if resultDegrees < 0 {
		return 360 + resultDegrees
	}
	return resultDegrees
}

type Track struct {
	Points        []gpx.TrackPoint
	DistanceKms   float64
	ElevationGain float64
	ElevationLoss float64
}

func (t Track) String() string {
	return fmt.Sprintf(
		"%.2f (#Points - %d) Gain: %.0fm Loss: %.0fm",
		t.DistanceKms,
		len(t.Points),
		t.ElevationGain,
		t.ElevationLoss,
	)
}

func Load(filename string) (*Track, error) {
	gpsTrack, err := Read(filename)
	if err != nil {
		return nil, fmt.Errorf(
			"couldn't read gps file: %s - %v", strconv.Quote(filename), err,
		)
	}

	return Data(gpsTrack), nil
}
