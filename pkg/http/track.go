package http

import (
	"fmt"

	"github.com/AngelVI13/satvar/pkg/gps"
)

func (s *Server) LoadTrack(filename string) error {
	track, err := gps.Load(filename)
	if err != nil {
		return fmt.Errorf("failed to load track: %v", err)
	}

	s.track = track
	return nil
}

func (s *Server) Track() *gps.Track {
	return s.track
}

func (s *Server) TrackLoaded() bool {
	return s.track != nil
}

func (s *Server) Location() *gps.Location {
	return s.location
}

func (s *Server) SetLocation(long, lat float64) {
	s.location = &gps.Location{
		Longitude: long,
		Latitude:  lat,
	}
}
