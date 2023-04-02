package drawing

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"math/rand"
	"os"

	"github.com/AngelVI13/satvar/pkg/gps"
	svg "github.com/ajstarks/svgo"
	"github.com/disintegration/imaging"
	gpx "github.com/sudhanshuraheja/go-garmin-gpx"
)

func CreateMapImage(
	track *gps.Track,
	userLocation *gps.Location,
	filename string,
) error {
	mapPoints, _, width, height := mapData(track.Points, userLocation)
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	drawRoute(img, mapPoints)

	flippedImg := imaging.FlipV(img.SubImage(img.Bounds()))

	// Encode as PNG.
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	err = png.Encode(f, flippedImg)
	return err
}

func CreateMapImageSvg(
	track *gps.Track,
	userLocation *gps.Location,
	direction float64,
) []byte {
	mapPoints, user, width, height := mapData(track.Points, userLocation)
	return drawRouteSvg(mapPoints, user, width, height, direction)
}

func calculateViewBox(width, height int, userLoc *MapPoint, followUser bool) string {
	if !followUser || userLoc == nil {
		// TODO: use this to show the map in full-screen
		return fmt.Sprintf("viewBox=\"0 0 %d %d\"", width, height)
	}

	// Adjust viewBox to center user on screen
	// TODO: get current screensize and adjust padding accordingly
	paddingX := 600
	paddingY := 600

	var (
		startX int
		startY int
	)

	startX = userLoc.x - paddingX
	startY = height - userLoc.y - paddingY

	return fmt.Sprintf("viewBox=\"%d %d %d %d\"", startX, startY, 2*paddingX, 2*paddingY)
}

func drawRouteSvg(
	points []MapPoint,
	userLoc *MapPoint,
	width, height int,
	direction float64,
) []byte {
	startEndCircleSize := 10
	chunkSize := 5

	xPointsToDraw := make([]int, 0, len(points)/chunkSize)
	yPointsToDraw := make([]int, 0, len(points)/chunkSize)

	for idx := range points {
		if idx%chunkSize == 0 {
			xPointsToDraw = append(xPointsToDraw, points[idx].x)
			yPointsToDraw = append(yPointsToDraw, height-points[idx].y)
		}
	}

	if len(xPointsToDraw) < 1 {
		return nil
	}

	var buf bytes.Buffer
	s := svg.New(&buf)

	followUser := false
	viewBox := calculateViewBox(width, height, userLoc, followUser)
	preserveAspectRatio := "preserveAspectRatio=\"xMinYMin meet\""
	rotation := rotate(direction+270, userLoc.x, height-userLoc.y)
	s.Startpercent(100, 100, viewBox, preserveAspectRatio, rotation)

	// draw start circle
	startPointX := xPointsToDraw[0]
	startPointY := yPointsToDraw[0]
	s.Circle(startPointX, startPointY, startEndCircleSize, "fill:blue")

	// Draw a polyline between every `chunkSize` points
	s.Polyline(
		xPointsToDraw,
		yPointsToDraw,
		"fill:none;stroke-width:2; stroke:black",
	)

	// draw finish circle
	endPointX := xPointsToDraw[len(xPointsToDraw)-1]
	endPointY := yPointsToDraw[len(yPointsToDraw)-1]

	distanceBetweenArrows := chunkSize * 20

	// draw direction arrows
	for i := distanceBetweenArrows; i < len(xPointsToDraw); i += distanceBetweenArrows {
		imageSize := startEndCircleSize * 2

		imageX := xPointsToDraw[i]
		imageY := yPointsToDraw[i]

		prevX := xPointsToDraw[i-chunkSize]
		prevY := yPointsToDraw[i-chunkSize]

		imageAngle := gps.Angle(prevX, prevY, imageX, imageY)

		transform := rotate(-imageAngle, imageX, imageY)

		emptyclose := "/>\n"
		imageSvg := fmt.Sprintf(
			`<image %s %s %s %s`,
			dim(imageX-imageSize, imageY-imageSize/2, imageSize, imageSize),
			href("assets/arrow_s.png"),
			transform,
			emptyclose,
		)

		buf.WriteString(imageSvg)
	}

	// draw finish
	s.Circle(endPointX, endPointY, startEndCircleSize, "fill:red")

	// draw user
	if userLoc != nil {
		log.Println(userLoc.x, height-userLoc.y, direction)
		s.Circle(userLoc.x, height-userLoc.y, startEndCircleSize, "fill:green")
	}

	s.Gend()
	s.End()

	return buf.Bytes()
}

func rotate(deg float64, x, y int) string {
	return fmt.Sprintf("transform=\"rotate(%f, %d, %d)\"", deg, x, y)
}

// SVG funcs
// href returns the href name and attribute
func href(s string) string { return fmt.Sprintf(`xlink:href="%s"`, s) }

// dim returns the dimension string (x, y coordinates and width, height)
func dim(x int, y int, w int, h int) string {
	return fmt.Sprintf(`x="%d" y="%d" width="%d" height="%d"`, x, y, w, h)
}

type MapPoint struct {
	x   int
	y   int
	geo *gpx.TrackPoint
}

// https://en.wikipedia.org/wiki/Decimal_degrees
// 10_000 for 11.1m accuracy (best for testing)
// 100_000 for 1.1m accuracy
// 1_000_000 for 1.1cm accuracy
const CoordScale = 10_000

func mapPoint(degrees float64) int {
	return int(math.Round(CoordScale * degrees))
}

func mapData(
	points []gpx.TrackPoint,
	userLoc *gps.Location,
) (mapPoints []MapPoint, userLocation *MapPoint, width, height int) {
	var (
		minLat  = 360.0
		minLong = 360.0
		maxLat  = 0.0
		maxLong = 0.0
	)

	for _, point := range points {
		if float64(point.Latitude) < minLat {
			minLat = float64(point.Latitude)
		}
		if float64(point.Latitude) > maxLat {
			maxLat = float64(point.Latitude)
		}

		if float64(point.Longitude) < minLong {
			minLong = float64(point.Longitude)
		}
		if float64(point.Longitude) > maxLong {
			maxLong = float64(point.Longitude)
		}
	}

	height = mapPoint(maxLat - minLat)
	width = mapPoint(maxLong - minLong)

	// TODO: return heigh and width scaling functions and do that during drawing

	for _, point := range points {
		geoPoint := &point
		mapPoint := MapPoint{
			x: scaleMapPoint(
				float64(point.Longitude),
				minLong,
				maxLong,
				width,
			),
			y:   scaleMapPoint(float64(point.Latitude), minLat, maxLat, height),
			geo: geoPoint,
		}
		mapPoints = append(mapPoints, mapPoint)
	}

	if userLoc != nil {
		userLocation = &MapPoint{
			x:   scaleMapPoint(userLoc.Longitude, minLong, maxLong, width),
			y:   scaleMapPoint(userLoc.Latitude, minLat, maxLat, height),
			geo: nil,
		}
	}

	return mapPoints, userLocation, width, height
}

func scaleMapPoint(x, minX, maxX float64, toSize int) int {
	normalized := normalize(x, minX, maxX)
	scaled := normalized * float64(toSize)
	rounded := math.Round(scaled)
	return int(rounded)
}

func normalize(x, minX, maxX float64) float64 {
	return (x - minX) / (maxX - minX)
}

var (
	Purple = color.RGBA{0x71, 0x03, 0x8A, 0xFF}
	Black  = color.RGBA{0x00, 0x00, 0x00, 0xFF}
	Green  = color.RGBA{0x00, 0xFF, 0x00, 0xFF}
	Blue   = color.RGBA{0x00, 0x00, 0xFF, 0xFF}
	Red    = color.RGBA{0xFF, 0x00, 0x00, 0xFF}
	White  = color.RGBA{0xd3, 0xd3, 0xd3, 0xFF}
)

var Colors = [...]color.RGBA{Purple, Black, Green, Red, Blue}

func drawRoute(img *image.RGBA, points []MapPoint) {
	size := img.Bounds().Size()
	width := size.X
	height := size.Y

	randomIndex := rand.Intn(len(Colors))
	color := Colors[randomIndex]

	// Set color for each pixel.
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, White)
		}
	}

	for _, point := range points {
		img.Set(point.x, point.y, color)
		// img.Set(point.x, point.y, Purple)
	}
}
