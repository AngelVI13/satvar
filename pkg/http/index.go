package http

import (
	"html/template"

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

	svgBytes := drawing.CreateMapImageSvg(s.Track())
	data["SvgImage"] = template.HTML(svgBytes)

	return c.Render(IndexView, data)
}
