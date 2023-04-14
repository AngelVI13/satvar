package http

import (
	"embed"
	"fmt"
	"log"
	"net/http"

	"github.com/AngelVI13/satvar/pkg/gps"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/template/html"
)

const (
	IndexView      = "views/index"
	MainLayoutView = "views/layouts/main"
	IndexUrl       = "/"
	MapUrl         = "/map"
)

var MapUrlFull = fmt.Sprintf("%s/:lat/:long/:sWidth/:sHeight", MapUrl)

type Server struct {
	*fiber.App
	sessionStore *session.Store
	track        *gps.Track
	trackFile    string
	location     map[string]*gps.Location
	route        map[string]*gps.Route
	debug        bool
}

// Generate png image from gps points & current location.
// Set the generated image as an html element to be displayed.
// In JS obtain location once per second and call backend which
// regenerates image and refreshes display
func NewServer(viewsfs embed.FS, session *session.Store, debug bool) *Server {
	s := Server{
		sessionStore: session,
		location:     make(map[string]*gps.Location),
		route:        make(map[string]*gps.Route),
		debug:        debug,
	}

	engine := html.NewFileSystem(http.FS(viewsfs), ".html")

	// TODO: maybe we can use template funcs to provide location to backend
	// templatefuncs.Register(db, engine)

	// Pass the engine to the Views
	app := fiber.New(fiber.Config{
		Views:       engine,
		ViewsLayout: MainLayoutView,
		BodyLimit:   50 * 1024 * 1024, // 50 MB
	})

	s.App = app

	// Middleware
	app.Use("/css", embededFileServer(viewsfs, "views/static/css"))
	app.Static("/", "./views/static")
	app.Use(loggingHandler)

	// index
	app.Get(
		IndexUrl,
		s.HandleIndex,
	)

	app.Get(
		MapUrlFull,
		s.HandleMap,
	)

	return &s
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

func embededFileServer(viewsfs embed.FS, path string) fiber.Handler {
	return filesystem.New(filesystem.Config{
		Root:       http.FS(viewsfs),
		PathPrefix: path,
		Browse:     true,
	})
}
