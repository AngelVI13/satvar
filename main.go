package main

import (
	"embed"
	"log"
	"os"

	"github.com/AngelVI13/satvar/pkg/http"
	"github.com/gofiber/fiber/v2/middleware/session"
)

//go:embed views/*
var viewsfs embed.FS

func main() {
	// Initialize default config
	// This stores all of your app's sessions
	store := session.New()

	_, debug := os.LookupEnv("DEBUG")
	server := http.NewServer(viewsfs, store, debug)

	log.Fatal(server.Listen(":5000"))
}
