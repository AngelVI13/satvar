package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net/http"

	geo "github.com/marcinwyszynski/geopoint"
	gpx "github.com/sudhanshuraheja/go-garmin-gpx"
)

// TODO: visualize gps coordinates on map:
// https://www.here.com/learn/blog/reverse-geocoding-a-location-using-golang
func main() {
	file := "Vilnius100km.gpx"
	g, err := gpx.ParseFile(file)

	if err != nil {
		log.Fatal(err)
	}

	log.Println(len(g.Tracks))
	if len(g.Tracks) != 1 {
		log.Fatalf(
			"currently only 1 track per file is supported: got %d",
			len(g.Tracks),
		)
	}

	track := g.Tracks[0]

	var (
		kms             geo.Kilometres = 0.0
		geoPoints       []*geo.GeoPoint
		elevationPoints []float64
	)

	for _, segment := range track.TrackSegments {
		for _, point := range segment.TrackPoint {
			elevationPoints = append(elevationPoints, point.Elevation)

			gPoint := geo.NewGeoPoint(geo.Degrees(point.Latitude), geo.Degrees(point.Longitude))
			geoPoints = append(geoPoints, gPoint)

			if len(geoPoints) <= 1 {
				continue
			}

			kms += gPoint.DistanceTo(geoPoints[len(geoPoints)-2], geo.Haversine)

		}
	}

	log.Println("Kilometres", kms)
	log.Println("#Points", len(elevationPoints))

	elevationGain, elevationLoss := calculateElevation(elevationPoints, 60)
	log.Println("Gain", elevationGain)
	log.Println("Loss", elevationLoss)
	// log.Println(elevationPoints)

	/*
		mux := mux.NewRouter()

		mux.HandleFunc("/", home)
		mux.HandleFunc("/location/{lat}/{long}", location)

		http.ListenAndServe(":8080", mux)
	*/
}

func calculateElevation(points []float64, chunkSize int) (gain, loss float64) {
	var temp = make([]float64, chunkSize)
	var avg []float64

	for i := 0; i < len(points); i++ {
		temp[i%chunkSize] = points[i]

		if i%chunkSize == 0 {
			avgElev := 0.0
			for _, elev := range temp {
				avgElev += elev
			}
			avgElev /= float64(chunkSize)
			avg = append(avg, avgElev)
		}
	}

	for i := 1; i < len(avg); i++ {
		diff := avg[i] - avg[i-1]

		if diff > 0 {
			gain += diff
		} else if diff < 0 {
			loss += (diff * -1.0)
		}
	}
	return gain, loss
}

func home(w http.ResponseWriter, r *http.Request) {

	var templates = template.Must(template.New("geolocate").ParseFiles("getlocation.html"))

	err := templates.ExecuteTemplate(w, "getlocation.html", nil)

	if err != nil {
		panic(err)
	}

	prompt := "Detecting your location. Please click 'Allow' button"
	w.Write([]byte(prompt))

}

func location(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	lat := vars["lat"]
	long := vars["long"]

	w.Write([]byte(fmt.Sprintf("Lat is %s \n", lat)))

	w.Write([]byte(fmt.Sprintf("Long is %s \n", long)))

	fmt.Printf("Lat is %s and Long is %s \n", lat, long)

	// if you want to get timezone from latitude and longitude
	// checkout http://www.geonames.org/export/web-services.html#timezone

}
