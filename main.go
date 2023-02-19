package main

import (
	"embed"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
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

var (
	Kms             geo.Kilometres = 0.0
	GeoPoints       []*geo.GeoPoint
	ElevationPoints []float64
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

	for _, segment := range track.TrackSegments {
		for _, point := range segment.TrackPoint {
			ElevationPoints = append(ElevationPoints, point.Elevation)

			gPoint := geo.NewGeoPoint(geo.Degrees(point.Latitude), geo.Degrees(point.Longitude))
			GeoPoints = append(GeoPoints, gPoint)

			if len(GeoPoints) <= 1 {
				continue
			}

			Kms += gPoint.DistanceTo(GeoPoints[len(GeoPoints)-2], geo.Haversine)

		}
	}

	log.Println("Kilometres", Kms)
	log.Println("#Points", len(ElevationPoints))

	// TODO: determine chunkSize automatically based on number of elevation points
	elevationGain, elevationLoss := calculateElevation(ElevationPoints, 60)
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
	app.Use("/assets", assetsFileServer()) // TODO: temporary remove later
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

	// TODO: currently graph does not look good - fix it
	path, err := createGraph(GeoPoints)
	if err != nil {
		flash.WithError(c, flashMessage(fmt.Sprintf(
			"error creating gps graph: %v", err), LevelPrimary))
		return c.Render(IndexView, data)
	}

	data["Image"] = "assets/track_1.png"
	log.Println(path)
	// TODO: pass graph html to template via `data`

	return c.Render(IndexView, data)
}

func HandleLocation(c *fiber.Ctx) error {
	resetQueryString(c)

	longitude := c.Params("long")
	latitude := c.Params("lat")
	log.Println(longitude, latitude)

	return c.RedirectBack(IndexUrl)
}

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

func assetsFileServer() fiber.Handler {
	return filesystem.New(filesystem.Config{
		Root: http.FS(viewsfs),
		// TODO: This is hardcoded
		PathPrefix: "views/static/assets",
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

func createGraph(points []*geo.GeoPoint) (path string, err error) {
	path = "scatter.html"

	page := components.NewPage()
	page.AddCharts(
		scatterBase(points),
	)
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}

	page.Render(io.MultiWriter(f))
	return path, nil
}

func scatterBase(points []*geo.GeoPoint) *charts.Scatter {
	scatter := charts.NewScatter()
	scatter.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "basic scatter example"}),
	)

	numberOfPoints := len(points)
	xAxis := make([]geo.Degrees, numberOfPoints)
	scatterItems := make([]opts.ScatterData, numberOfPoints)

	for _, point := range points {
		xAxis = append(xAxis, point.Longitude)
		scatterItems = append(scatterItems, opts.ScatterData{
			Value: point.Latitude,
			// NOTE: can also use "arrow" but have to compute angel of rotation
			Symbol:       "circle",
			SymbolSize:   20,
			SymbolRotate: 10,
		})
	}

	scatter.SetXAxis(xAxis).AddSeries("Points", scatterItems)

	return scatter
}
