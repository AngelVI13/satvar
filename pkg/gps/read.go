package gps

import (
	"fmt"
	"log"

	gpx "github.com/sudhanshuraheja/go-garmin-gpx"
)

func Read(path string) (*gpx.Track, error) {
	g, err := gpx.ParseFile(path)

	if err != nil {
		log.Fatal(err)
	}

	if len(g.Tracks) != 1 {
		return nil, fmt.Errorf(
			"currently only 1 track per file is supported: got %d",
			len(g.Tracks),
		)
	}

	track := g.Tracks[0]
	return &track, nil
}

func Data(track *gpx.Track) *Track {
	t := &Track{}

	for _, segment := range track.TrackSegments {
		t.Points = append(t.Points, segment.TrackPoint...)
	}

	// TODO: determine chunkSize automatically based on number of gpx points
	distanceKms, elevationGain, elevationLoss := calculateProperties(t.Points, 60)

	t.DistanceKms = distanceKms
	t.ElevationGain = elevationGain
	t.ElevationLoss = elevationLoss

	return t
}
