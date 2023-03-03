package http

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

func (s *Server) HandleLocation(c *fiber.Ctx) error {
	longitude := c.Params("long")
	latitude := c.Params("lat")
	log.Println(longitude, latitude)
	return nil
}
