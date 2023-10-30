package main

import (
	"embed"
	"io/fs"
	"net/http"
	"os"

	"github.com/Masterminds/sprig/v3"
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

var config Config

type Config struct {
	BrevoKey    string `mapstructure:"AUTH_BREVO_APIKEY"`
	SendgridKey string `mapstructure:"AUTH_SENDGRID_APIKEY"`
	AdminEmail  string `mapstructure:"AUTH_ADMIN_EMAIL"`
	LogLevel    string `mapstructure:"AUTH_LOGLEVEL"`
	Env         string `mapstructure:"AUTH_ENV"`
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	viper.SetDefault("Env", "Dev")
	viper.SetEnvPrefix("auth")
	viper.AutomaticEnv()
	viper.SetConfigFile("dev.env")
	err := viper.ReadInConfig()
	if err != nil {
		log.Warn().Msgf("Failed to read config: %v", err)
	}
	err = viper.Unmarshal(&config)
	if err != nil {
		log.Err(err)
	}
	level, err := zerolog.ParseLevel(config.LogLevel)
	if err != nil {
		log.Err(err)
		if viper.Get("Env") == "Dev" {
			log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		}

		zerolog.SetGlobalLevel(level)

	}

	db.Init()
	userDb := db.InitUsers()
	files, err := fs.Sub(viewsFilesystem, "views")
	if err != nil {
		log.Error().Msg("Failed to open subdir of filesystem")
	}
	httpViews := http.FS(files)

	engine := html.NewFileSystem(httpViews, ".html")
	if viper.Get("Env") == "Dev" {
		engine.Debug(true)
	}
	engine.AddFuncMap(sprig.FuncMap())

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
			log.Print(err)
			log.Printf("login status: %v", status)
			return c.Render("login", status)
		}
		status.Username = username
		status.Status = true

		return c.Render("login", status)
	})

	api.RegisterHXRoutes(app.Group("/hx"))

	api.RegisterAuthRoutes(app.Group("/auth"), &userDb)

	app.Use(api.NewLoginRedirect())

	api.RegisterUserRoutes(app.Group("/api/users"), &userDb)

	app.Get("/", func(c *fiber.Ctx) error {
		users := userDb.GetUsers()
		for _, v := range users {
			log.Print(string(v.ID))
			log.Print(v.Status)
			log.Print(v.Username)
			log.Print(v.Role)
		}
		return c.Render("layout", struct {
			Accounts []db.User
			Title    string
		}{users, "Manage Accounts"})
	})

	log.Fatal().Err(app.Listen(":4200"))
}
