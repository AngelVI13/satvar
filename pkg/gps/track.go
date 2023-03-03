package gps

import (
	"fmt"
	"strconv"

	gpx "github.com/sudhanshuraheja/go-garmin-gpx"
)

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
			"Couldn't read gps file: %s - %v", strconv.Quote(filename), err,
		)
	}

	return Data(gpsTrack), nil
}
