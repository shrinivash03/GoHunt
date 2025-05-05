package main

import (
	"ash/gohunt/db"
	"ash/gohunt/routes"
	"ash/gohunt/utils"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/joho/godotenv"
)

func main() {
	env := godotenv.Load()
	if env != nil {
		fmt.Println("cannot find environment variable")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080"
	} else {
		port = ":" + port
	}

	app := fiber.New(fiber.Config{
		IdleTimeout: 5 * time.Second,
	})

	app.Use(compress.New())
	db.InitDB()
	routes.SetRoutes(app)
	utils.StartCronJobs()

	// Start our server and listen for shutdown
	go func() {
		if err := app.Listen(port); err != nil {
			log.Panic(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c // Block the main thread until interrupted
	app.Shutdown()
	fmt.Println("shutting down server")
}
