package http

import (
	"fmt"
	"html/template"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sujit-baniya/flash"
)

const mapFilename = "Vilnius100km.gpx"

func (s *Server) HandleIndex(c *fiber.Ctx) error {
	userId, err := s.userId(c)
	if err != nil {
		return err
	}

	if !s.TrackLoaded(mapFilename) {
		s.LoadTrack(mapFilename)
	}

	longitude := c.Query("long")
	latitude := c.Query("lat")
	s.processLocation(userId, longitude, latitude)

	screenWidth := c.Query("sWidth")
	screenHeight := c.Query("sHeight")
	s.processScreenSize(userId, screenWidth, screenHeight)

	data := flash.Get(c)
	data["Title"] = "Satvar"

	svgBytes, err := s.GenerateMap(mapFilename, userId)
	if err != nil {
		flash.WithError(c, flashMessage(err.Error(), LevelDanger))
		return c.Render(IndexView, data)
	}

	data["SvgImage"] = template.HTML(svgBytes)

	return c.Render(IndexView, data)
}

// only used for debugging
var currentTrackIdx = map[string]int{}

func (s *Server) processLocation(id string, longitude, latitude string) {
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
		log.Println(currentTrackIdx)

		// TODO: store progress of demo based on user ID
		track := s.Track()
		point := track.Points[currentTrackIdx[id]]

		currentTrackIdx[id] += 5
		if currentTrackIdx[id] >= len(track.Points) {
			currentTrackIdx[id] = 0
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

	s.SetLocation(id, longitudeFloat, latitudeFloat)
}

func (s *Server) processScreenSize(id, width, height string) {
}

func (s *Server) userId(c *fiber.Ctx) (string, error) {
	sess, err := s.sessionStore.Get(c)
	if err != nil {
		return "", err
	}

	// Get value
	id := sess.Get("user_id")
	if id == nil {
		id = uuid.New().String()
		sess.Set("user_id", id)

		// Save session
		if err := sess.Save(); err != nil {
			return "", err
		}
	}
	userId, ok := id.(string)
	if !ok {
		return "", fmt.Errorf("failed to cast user id %v to string", id)
	}

	return userId, nil
}
