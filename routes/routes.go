package routes

import (
	"time"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/cache"
)

func render(c *fiber.Ctx, component templ.Component, options ...func(*templ.ComponentHandler)) error {
	componentHandler := templ.Handler(component)
	for _, o := range options {
		o(componentHandler)
	}
	return adaptor.HTTPHandler(componentHandler)(c)
}

func SetRoutes(app *fiber.App) {
	app.Get("/", AuthMiddleWare, DashboardHandler)

	app.Post("/", AuthMiddleWare, DashboardPostHandler)

	app.Get("/login", LoginHandler)

	app.Post("/login", LoginPostHandler)

	app.Get("/register", RegisterHandler)

	app.Post("/register", RegisterPostHandler)
	
	app.Post("/logout", LogoutHandler)

	app.Post("/search", HandleSearch)
	app.Use("/search", cache.New(cache.Config{
		Next: func(c * fiber.Ctx) bool {
			return c.Query("noCache") == "true"
		},
		Expiration: 30 * time.Minute,
		CacheControl: true,
	}))
}
