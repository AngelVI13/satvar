package http

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

type MessageLevel string

const (
	LevelPrimary MessageLevel = "primary"
	LevelSuccess MessageLevel = "success"
	LevelWarning MessageLevel = "warning"
	LevelDanger  MessageLevel = "danger"
)

func flashMessage(message string, level MessageLevel) fiber.Map {
	log.Println(message)
	return fiber.Map{
		"Message": message,
		"Level":   level,
	}
}
