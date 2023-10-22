package main

import (
	"embed"
	"flag"
	"io/fs"
	"net/http"

	"github.com/a19simma/vanilla-js/api"
	"github.com/a19simma/vanilla-js/pkg/db"
	"github.com/a19simma/vanilla-js/pkg/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/template/html/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

//go:embed views/*
var viewsFilesystem embed.FS

//go:embed dist
var dist embed.FS

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	debug := flag.Bool("debug", true, "sets log level to debug")

	flag.Parse()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	viper.SetConfigFile(".env")
	err := viper.ReadInConfig()
	if err != nil {
		log.Error().Msgf("Failed to read config: %v", err)
	}

	db.Init()
	db.InitUsers()
	files, err := fs.Sub(viewsFilesystem, "views")
	if err != nil {
		log.Error().Msg("Failed to open subdir of filesystem")
	}
	httpViews := http.FS(files)

	engine := html.NewFileSystem(httpViews, ".html")
	engine.Debug(true)
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Use(middleware.NewLoggerMiddleWare())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "*",
	}))

	dist, err := fs.Sub(dist, "dist")
	if err != nil {
		log.Err(err)
	}

	assets, err := fs.Sub(dist, "assets")
	if err != nil {
		log.Err(err)
	}
	scripts, err := fs.Sub(dist, "scripts")
	if err != nil {
		log.Err(err)
	}
	styles, err := fs.Sub(dist, "styles")
	if err != nil {
		log.Err(err)
	}

	app.Use("/assets", filesystem.New(filesystem.Config{
		Root: http.FS(assets),
	}))
	app.Use("/scripts", filesystem.New(filesystem.Config{
		Root: http.FS(scripts),
	}))
	app.Use("/styles", filesystem.New(filesystem.Config{
		Root: http.FS(styles),
	}))

	app.Get("/login", func(c *fiber.Ctx) error {
		username, err := api.CheckLoginStatus(c)
		status := struct {
			Status   bool
			Username string
		}{
			Status: false,
		}
		if err != nil {
			return c.Render("login", status)
		}
		status.Username = username
		status.Status = true

		return c.Render("login", status)
	})

	api.RegisterHXRoutes(app.Group("/hx"))

	api.RegisterAuthRoutes(app.Group("/auth"))

	app.Use(api.NewLoginRedirect())

	app.Get("/", func(c *fiber.Ctx) error { return c.Render("index", nil) })

	log.Fatal().Err(app.Listen(":4200"))
}
