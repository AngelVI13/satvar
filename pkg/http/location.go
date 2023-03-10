package http

import (
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func (s *Server) HandleLocation(c *fiber.Ctx) error {
	longitude := c.Params("long")
	latitude := c.Params("lat")
	log.Println(longitude, latitude)

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

	s.SetLocation(longitudeFloat, latitudeFloat)
	return nil
}
