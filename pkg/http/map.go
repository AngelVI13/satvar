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

	svgBytes, err := s.GenerateMap(mapFilename)
	if err != nil {
		return c.SendStatus(501)
	}
	return c.SendString(string(svgBytes))
}
