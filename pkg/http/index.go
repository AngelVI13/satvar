package http

import (
	"html/template"
	"log"
	"strconv"

	"github.com/AngelVI13/satvar/pkg/drawing"
	"github.com/gofiber/fiber/v2"
	"github.com/sujit-baniya/flash"
)

var myNumber int

func (s *Server) HandleIndex(c *fiber.Ctx) error {
	longitude := c.Query("long")
	latitude := c.Query("lat")
	s.processLocation(longitude, latitude)

	screenWidth := c.Query("sWidth")
	screenHeight := c.Query("sHeight")
	s.processScreenSize(screenWidth, screenHeight)

	data := flash.Get(c)
	data["Title"] = "Satvar"

	if !s.TrackLoaded() {
		// TODO: visualize gps coordinates on map:
		// https://www.here.com/learn/blog/reverse-geocoding-a-location-using-golang
		filename := "Vilnius100km.gpx"
		// filename := "home_run.gpx"
		err := s.LoadTrack(filename)
		if err != nil {
			flash.WithError(c, flashMessage(err.Error(), LevelDanger))
			return c.Render(IndexView, data)
		}
	}

	svgBytes := drawing.CreateMapImageSvg(s.Track(), s.Location())
	data["SvgImage"] = template.HTML(svgBytes)

	return c.Render(IndexView, data)
}

// only used for debuggin
var currentTrackIdx int

func (s *Server) processLocation(longitude, latitude string) {
	if (longitude == "" || latitude == "") && !s.debug {
		return
	}

	var (
		longitudeFloat float64
		latitudeFloat  float64
	)

	if s.debug {
		// This imitates a person going through the course route
		if !s.TrackLoaded() {
			s.LoadTrack("Vilnius100km.gpx")
		}

		track := s.Track()
		point := track.Points[currentTrackIdx]

		currentTrackIdx += 500
		if currentTrackIdx >= len(track.Points) {
			currentTrackIdx = 0
		}

		longitudeFloat = float64(point.Longitude)
		latitudeFloat = float64(point.Latitude)
	} else {
		var err error
		longitudeFloat, err = strconv.ParseFloat(longitude, 64)
		if err != nil {
			log.Fatalf(
				"error while converting longitude %s to float64: %v",
				strconv.Quote(longitude),
				err,
			)
		}

		latitudeFloat, err = strconv.ParseFloat(latitude, 64)
		if err != nil {
			log.Fatalf(
				"error while converting latitude %s to float64: %v",
				strconv.Quote(latitude),
				err,
			)
		}
	}

	s.SetLocation(longitudeFloat, latitudeFloat)
}

func (s *Server) processScreenSize(width, height string) {
	log.Println(width, height)
}
