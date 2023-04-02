package http

import (
	"fmt"

	"github.com/AngelVI13/satvar/pkg/drawing"
	"github.com/AngelVI13/satvar/pkg/gps"
)

func (s *Server) LoadTrack(filename string) error {
	track, err := gps.Load(filename)
	if err != nil {
		return fmt.Errorf("failed to load track: %v", err)
	}

	s.track = track
	s.trackFile = filename
	// TODO: this should be per session
	s.route = gps.NewRoute()

	return nil
}

func (s *Server) Track() *gps.Track {
	return s.track
}

func (s *Server) TrackLoaded(filename string) bool {
	return s.track != nil && s.trackFile == filename
}

func (s *Server) Location() *gps.Location {
	return s.location
}

func (s *Server) SetLocation(long, lat float64) {
	s.location = &gps.Location{
		Longitude: long,
		Latitude:  lat,
	}
	s.route.AddPoint(*s.location)
}

func (s *Server) Direction() float64 {
	return s.route.Direction()
}

func (s *Server) GenerateMap(filename string) ([]byte, error) {
	if !s.TrackLoaded(filename) {
		// TODO: visualize gps coordinates on map:
		// https://www.here.com/learn/blog/reverse-geocoding-a-location-using-golang
		err := s.LoadTrack(filename)
		if err != nil {
			return nil, err
		}
	}

	svgBytes := drawing.CreateMapImageSvg(s.Track(), s.Location(), s.Direction())
	return svgBytes, nil
}
