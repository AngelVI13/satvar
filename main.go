package main

import (
	"embed"
	"log"
	"os"

	"github.com/AngelVI13/satvar/pkg/http"
)

//go:embed views/*
var viewsfs embed.FS

func main() {
	_, debug := os.LookupEnv("DEBUG")
	server := http.NewServer(viewsfs, debug)

	log.Fatal(server.Listen(":5000"))
}
