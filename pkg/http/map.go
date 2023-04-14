package http

import (
	"github.com/gofiber/fiber/v2"
)

func (s *Server) HandleMap(c *fiber.Ctx) error {
	userId, err := s.userId(c)
	if err != nil {
		return err
	}

	if !s.TrackLoaded(mapFilename) {
		s.LoadTrack(mapFilename)
	}
	longitude := c.Params("long")
	latitude := c.Params("lat")
	s.processLocation(userId, longitude, latitude)

	screenWidth := c.Params("sWidth")
	screenHeight := c.Params("sHeight")
	s.processScreenSize(userId, screenWidth, screenHeight)

	svgBytes, err := s.GenerateMap(mapFilename, userId)
	if err != nil {
		return c.SendStatus(501)
	}
	return c.SendString(string(svgBytes))
}
