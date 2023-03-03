package gps

import (
	geo "github.com/marcinwyszynski/geopoint"
	gpx "github.com/sudhanshuraheja/go-garmin-gpx"
)

// calculateProperties Calculates distance (km), elevation gain & loss given a slice of gpx points.
// Here chunkSize indicates the smoothing filter size (onlu used for elevation calculations).
// Since GPS tracks usually record at approx. once per second if you set a chunkSize of 60 -> each 60
// elevation points will be grouped together and one average value will be produced
// for them.
func calculateProperties(points []gpx.TrackPoint, chunkSize int) (
	distanceKms,
	elevGainMeters,
	elevLossMeters float64,
) {
	var (
		tempElev      = make([]float64, chunkSize)
		lastElevation float64
	)

	for i := 0; i < len(points); i++ {
		tempElev[i%chunkSize] = points[i].Elevation

		// elevation calculation with smoothing
		if i%chunkSize == 0 {
			avgElev := 0.0
			for _, elev := range tempElev {
				avgElev += elev
			}
			avgElev /= float64(chunkSize)

			// gain/loss calculation based on avg elev for chunk
			diff := avgElev - lastElevation
			if diff > 0 {
				elevGainMeters += diff
			} else if diff < 0 {
				elevLossMeters += (diff * -1.0)
			}

			// update elevation for last chunk
			lastElevation = avgElev
		}

		// distance calculation
		if i >= 1 {
			current := geoPoint(&points[i])
			prev := geoPoint(&points[i-1])
			distanceKms += float64(current.DistanceTo(prev, geo.Haversine))
		}
	}

	return distanceKms, elevGainMeters, elevLossMeters
}

func geoPoint(point *gpx.TrackPoint) *geo.GeoPoint {
	return geo.NewGeoPoint(geo.Degrees(point.Latitude), geo.Degrees(point.Longitude))
}
