package gps

import (
	"fmt"
	"log"

	geo "github.com/marcinwyszynski/geopoint"
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
		for _, point := range segment.TrackPoint {
			t.ElevationPoints = append(t.ElevationPoints, point.Elevation)

			gpsPoint := &point
			t.Points = append(t.Points, gpsPoint)

			if len(t.Points) <= 1 {
				continue
			}

			currentPoint := geo.NewGeoPoint(geo.Degrees(point.Latitude), geo.Degrees(point.Longitude))
			prev := t.Points[len(t.Points)-2]
			prevPoint := geo.NewGeoPoint(geo.Degrees(prev.Latitude), geo.Degrees(prev.Longitude))
			t.DistanceKms += float64(currentPoint.DistanceTo(prevPoint, geo.Haversine))
		}
	}

	// TODO: determine chunkSize automatically based on number of elevation points
	elevationGain, elevationLoss := calculateElevation(t.ElevationPoints, 60)

	t.ElevationGain = elevationGain
	t.ElevationLoss = elevationLoss

	return t
}
