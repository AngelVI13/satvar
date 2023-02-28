package main

import (
	"embed"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"math"
	"net/http"
	"os"

	"github.com/disintegration/imaging"
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
	// path, err := createGraph(GeoPoints)
	path := "views/static/assets/map.png"
	err := createMapImage(GeoPoints, path)
	if err != nil {
		flash.WithError(c, flashMessage(fmt.Sprintf(
			"error creating gps graph: %v", err), LevelPrimary))
		return c.Render(IndexView, data)
	}

	data["Image"] = "assets/map.png"
	log.Println(path)

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

func createMapImage(points []*geo.GeoPoint, filename string) error {
	mapPoints, width, height := mapData(points)
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	drawBoard(img, mapPoints)

	flippedImg := imaging.FlipV(img.SubImage(img.Bounds()))

	// Encode as PNG.
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	err = png.Encode(f, flippedImg)
	return err
}

type MapPoint struct {
	x   int
	y   int
	geo *geo.GeoPoint
}

// https://en.wikipedia.org/wiki/Decimal_degrees
// 10_000 for 11.1m accuracy (best for testing)
// 100_000 for 1.1m accuracy
// 1_000_000 for 1.1cm accuracy
const CoordScale = 10_000

func mapPoint(coord geo.Degrees) int {
	return int(math.Round(CoordScale * float64(coord)))
}

func mapData(points []*geo.GeoPoint) ([]*MapPoint, int, int) {
	var (
		minLat  geo.Degrees = 360.0
		minLong geo.Degrees = 360.0
		maxLat  geo.Degrees = 0.0
		maxLong geo.Degrees = 0.0

		mapPoints []*MapPoint
	)

	for _, point := range points {
		if point.Latitude < minLat {
			minLat = point.Latitude
		}
		if point.Latitude > maxLat {
			maxLat = point.Latitude
		}

		if point.Longitude < minLong {
			minLong = point.Longitude
		}
		if point.Longitude > maxLong {
			maxLong = point.Longitude
		}
	}

	height := mapPoint(maxLat - minLat)
	width := mapPoint(maxLong - minLong)

	for _, point := range points {
		geoPoint := point
		mapPoint := &MapPoint{
			x:   scaleMapPoint(point.Longitude, minLong, maxLong, width),
			y:   scaleMapPoint(point.Latitude, minLat, maxLat, height),
			geo: geoPoint,
		}
		mapPoints = append(mapPoints, mapPoint)
	}

	return mapPoints, width, height
}

func scaleMapPoint(x, minX, maxX geo.Degrees, toSize int) int {
	normalized := normalize(float64(x), float64(minX), float64(maxX))
	scaled := normalized * float64(toSize)
	rounded := math.Round(scaled)
	return int(rounded)

}

func normalize(x, minX, maxX float64) float64 {
	return (x - minX) / (maxX - minX)
}

var Purple = color.RGBA{0x71, 0x03, 0x8A, 0xFF}
var White = color.RGBA{0xd3, 0xd3, 0xd3, 0xFF}

func drawBoard(img *image.RGBA, points []*MapPoint) {
	size := img.Bounds().Size()
	width := size.X
	height := size.Y

	// Set color for each pixel.
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, White)
		}
	}

	for _, point := range points {
		img.Set(point.x, point.y, Purple)
	}
}

func createGraph(points []*geo.GeoPoint) (path string, err error) {
	path = "scatter.html"

	page := components.NewPage()
	page.AddCharts(
		scatterBase(points, 60),
	)
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}

	page.Render(io.MultiWriter(f))
	return path, nil
}

func scatterBase(points []*geo.GeoPoint, chunkSize int) *charts.Scatter {
	scatter := charts.NewScatter()
	scatter.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "basic scatter example"}),
	)

	numberOfPoints := len(points)
	xAxis := make([]geo.Degrees, numberOfPoints)
	scatterItems := make([]opts.ScatterData, numberOfPoints)

	for idx, point := range points {
		if idx%chunkSize != 0 {
			continue
		}
		xAxis = append(xAxis, point.Longitude)
		scatterItems = append(scatterItems, opts.ScatterData{
			Value: point.Latitude,
			// NOTE: can also use "arrow" but have to compute angel of rotation
			Symbol:       "circle",
			SymbolSize:   5,
			SymbolRotate: 2,
		})
	}

	scatter.SetXAxis(xAxis).AddSeries("Points", scatterItems)

	scatter.SetGlobalOptions(
		charts.WithXAxisOpts(opts.XAxis{
			SplitNumber: 20,
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Scale: true,
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:       "inside",
			Start:      50,
			End:        100,
			XAxisIndex: []int{0},
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:       "slider",
			Start:      50,
			End:        100,
			XAxisIndex: []int{0},
		}),
		charts.WithLegendOpts(opts.Legend{
			Show: true,
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Show: true,
		}),
	)

	return scatter
}
