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
	return nil
}

func (s *Server) Track() *gps.Track {
	return s.track
}

func (s *Server) TrackLoaded(filename string) bool {
	return s.track != nil && s.trackFile == filename
}

func (s *Server) Location(id string) *gps.Location {
	return s.location[id]
}

func (s *Server) SetLocation(id string, long, lat float64) {
	loc := &gps.Location{
		Longitude: long,
		Latitude:  lat,
	}
	s.location[id] = loc

	if _, exists := s.route[id]; !exists {
		s.route[id] = gps.NewRoute()
	}
	s.route[id].AddPoint(*loc)
}

func (s *Server) Direction(id string) float64 {
	return s.route[id].Direction()
}

func (s *Server) GenerateMap(filename, id string) ([]byte, error) {
	if !s.TrackLoaded(filename) {
		// TODO: visualize gps coordinates on map:
		// https://www.here.com/learn/blog/reverse-geocoding-a-location-using-golang
		err := s.LoadTrack(filename)
		if err != nil {
			return nil, err
		}
	}

	svgBytes := drawing.CreateMapImageSvg(s.Track(), s.Location(id), s.Direction(id))
	return svgBytes, nil
}
