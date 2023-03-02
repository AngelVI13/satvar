package http

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

func (s *Server) HandleLocation(c *fiber.Ctx) error {
	resetQueryString(c)

	longitude := c.Params("long")
	latitude := c.Params("lat")
	log.Println(longitude, latitude)

	return c.RedirectBack(IndexUrl)
}
