package main

import (
	"embed"
	"log"

	"github.com/AngelVI13/satvar/pkg/http"
)

//go:embed views/*
var viewsfs embed.FS

func main() {
	server := http.NewServer(viewsfs)

	log.Fatal(server.Listen(":5000"))
}
