package http

import (
	"fmt"
	"time"

	"github.com/AngelVI13/satvar/pkg/drawing"
	"github.com/gofiber/fiber/v2"
	"github.com/sujit-baniya/flash"
)

var myNumber int

func (s *Server) HandleIndex(c *fiber.Ctx) error {
	data := flash.Get(c)
	data["Title"] = "Satvar"

	if !s.TrackLoaded() {
		// TODO: visualize gps coordinates on map:
		// https://www.here.com/learn/blog/reverse-geocoding-a-location-using-golang
		filename := "Vilnius100km.gpx"
		err := s.LoadTrack(filename)
		if err != nil {
			flash.WithError(c, flashMessage(err.Error(), LevelDanger))
			return c.Render(IndexView, data)
		}
	}

	// TODO: currently graph does not look good - fix it
	// path, err := createGraph(GeoPoints)
	base := "views/static/"

	timestamp := time.Now()
	mapFile := fmt.Sprintf("assets/map_%d.png", timestamp.UnixMilli())
	path := base + mapFile
	err := drawing.CreateMapImage(s.Track(), path)
	if err != nil {
		flash.WithError(c, flashMessage(fmt.Sprintf(
			"error creating gps graph: %v", err), LevelPrimary))
		return c.Render(IndexView, data)
	}

	data["Image"] = mapFile

	return c.Render(IndexView, data)
}
