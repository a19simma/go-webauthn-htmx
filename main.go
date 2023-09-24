package main

import (
	"log"

	"github.com/a19simma/vanilla-js/api"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

type Todo struct {
	Title string
	Done  bool
}

type TodoPageData struct {
	Title string
	Todos []Todo
}

func main() {
	engine := html.New("./views", ".html")
	engine.Debug(true)
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	todoData := TodoPageData{
		Title: "Hello todos",
		Todos: []Todo{
			{Title: "1", Done: false},
		},
	}

	app.Static("/public", "./public")
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("layout", todoData)
	})

	hx := app.Group("/hx")
	api.RegisterHXRoutes(hx)

	log.Fatal(app.Listen(":3000"))
}
