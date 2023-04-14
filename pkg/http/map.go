package http

import (
	"github.com/gofiber/fiber/v2"
)

func (s *Server) HandleMap(c *fiber.Ctx) error {
	if !s.TrackLoaded(mapFilename) {
		s.LoadTrack(mapFilename)
	}
	longitude := c.Params("long")
	latitude := c.Params("lat")
	s.processLocation(longitude, latitude)

	screenWidth := c.Params("sWidth")
	screenHeight := c.Params("sHeight")
	s.processScreenSize(screenWidth, screenHeight)

	filename := "Vilnius100km.gpx"
	svgBytes, err := s.GenerateMap(filename)
	if err != nil {
		return c.SendStatus(501)
	}
	return c.SendString(string(svgBytes))
}
