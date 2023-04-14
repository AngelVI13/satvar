package http

import (
	"html/template"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sujit-baniya/flash"
)

const mapFilename = "Vilnius100km.gpx"

func (s *Server) HandleIndex(c *fiber.Ctx) error {
	if !s.TrackLoaded(mapFilename) {
		s.LoadTrack(mapFilename)
	}

	longitude := c.Query("long")
	latitude := c.Query("lat")
	s.processLocation(longitude, latitude)

	screenWidth := c.Query("sWidth")
	screenHeight := c.Query("sHeight")
	s.processScreenSize(screenWidth, screenHeight)

	data := flash.Get(c)
	data["Title"] = "Satvar"

	svgBytes, err := s.GenerateMap(mapFilename)
	if err != nil {
		flash.WithError(c, flashMessage(err.Error(), LevelDanger))
		return c.Render(IndexView, data)
	}

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
		if !s.TrackLoaded(mapFilename) {
			s.LoadTrack(mapFilename)
		}

		track := s.Track()
		point := track.Points[currentTrackIdx]

		currentTrackIdx += 5
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
}
