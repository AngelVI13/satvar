package http

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

// var currentTrackIdx int

func (s *Server) HandleLocation(c *fiber.Ctx) error {
	/*
		longitude := c.Params("long")
		latitude := c.Params("lat")
	*/
	if !s.TrackLoaded() {
		s.LoadTrack("Vilnius100km.gpx")
	}

	track := s.Track()
	point := track.Points[currentTrackIdx]

	currentTrackIdx += 10
	if currentTrackIdx >= len(track.Points) {
		currentTrackIdx = 0
	}

	longitudeFloat := float64(point.Longitude)
	latitudeFloat := float64(point.Latitude)
	log.Printf("Loc: %f, %f", longitudeFloat, latitudeFloat)
	/*
		longitude := "25.280135"
		latitude := "54.66068"
	*/

	/*
		longitudeFloat, err := strconv.ParseFloat(longitude, 64)
		if err != nil {
			log.Fatalf(
				"error while converting longitude %s to float64: %v",
				strconv.Quote(longitude),
				err,
			)
		}

		latitudeFloat, err := strconv.ParseFloat(latitude, 64)
		if err != nil {
			log.Fatalf(
				"error while converting latitude %s to float64: %v",
				strconv.Quote(latitude),
				err,
			)
		}
	*/

	s.SetLocation(longitudeFloat, latitudeFloat)

	return s.HandleIndex(c)
}
