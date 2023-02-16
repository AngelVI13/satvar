package main

import (
	"log"

	gpx "github.com/sudhanshuraheja/go-garmin-gpx"
)

// TODO: visualize gps coordinates on map:
// https://www.here.com/learn/blog/reverse-geocoding-a-location-using-golang
func main() {
	file := "Vilnius100km.gpx"
	g, err := gpx.ParseFile(file)

	if err != nil {
		log.Fatal(err)
	}

	log.Println(len(g.Tracks))
	if len(g.Tracks) != 1 {
		log.Fatalf(
			"currently only 1 track per file is supported: got %d",
			len(g.Tracks),
		)
	}

	track := g.Tracks[0]

	for _, segment := range track.TrackSegments {
		for _, point := range segment.TrackPoint {
			log.Println(point.Latitude, point.Longitude, point.Elevation)
		}
	}

	log.Println(len(track.TrackSegments))
}
