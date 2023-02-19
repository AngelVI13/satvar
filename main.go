package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/template/html"
	"github.com/sujit-baniya/flash"

	geo "github.com/marcinwyszynski/geopoint"
	gpx "github.com/sudhanshuraheja/go-garmin-gpx"
)

//go:embed views/*
var viewsfs embed.FS

const (
	IndexView      = "views/index"
	MainLayoutView = "views/layouts/main"
	IndexUrl       = "/"
	LocationUrl    = "/location"
)

var (
	LocationUrlFull = fmt.Sprintf("%s/:lat/:long", LocationUrl)
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

	// TODO: determine chunkSize automatically based on number of elevation points
	elevationGain, elevationLoss := calculateElevation(elevationPoints, 60)
	log.Println("Gain", elevationGain)
	log.Println("Loss", elevationLoss)

	engine := html.NewFileSystem(http.FS(viewsfs), ".html")

	// TODO: maybe we can use template funcs to provide location to backend
	// templatefuncs.Register(db, engine)

	// Pass the engine to the Views
	app := fiber.New(fiber.Config{
		Views:       engine,
		ViewsLayout: MainLayoutView,
		BodyLimit:   50 * 1024 * 1024, // 50 MB
	})

	// Middleware
	app.Use("/css", cssFileServer())
	app.Use(loggingHandler)

	// index
	app.Get(
		IndexUrl,
		HandleIndex,
	)

	// location
	app.Post(
		LocationUrlFull,
		HandleLocation,
	)

	log.Fatal(app.Listen(":5000"))

	// Generate png image from gps points & current location.
	// Set the generated image as an html element to be displayed.
	// In JS obtain location once per second and call backend which
	// regenerates image and refreshes display
}

func HandleIndex(c *fiber.Ctx) error {
	data := flash.Get(c)
	data["Title"] = "Satvar"

	return c.Render(IndexView, data)
}

func HandleLocation(c *fiber.Ctx) error {
	resetQueryString(c)

	longitude := c.Params("long")
	latitude := c.Params("lat")
	log.Println(longitude, latitude)

	return c.RedirectBack(IndexUrl)
}

func resetQueryString(c *fiber.Ctx) {
	c.Request().URI().SetQueryString("")
}

func loggingHandler(c *fiber.Ctx) error {
	uri := c.Request().URI().Path()
	method := c.Request().Header.Method()

	log.Printf("%s %s", method, uri)

	return c.Next()
}

func cssFileServer() fiber.Handler {
	return filesystem.New(filesystem.Config{
		Root: http.FS(viewsfs),
		// TODO: This is hardcoded
		PathPrefix: "views/static/css",
		Browse:     true,
	})
}

// calculateElevation Calculates elevation gain & loss give a slice of elevation data.
// Here chunkSize indicates the smoothing filter size. Since GPS tracks usually
// record at approx. once per second if you set a chunkSize of 60 -> each 60
// elevation points will be grouped together and one average value will be produced
// for them.
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
