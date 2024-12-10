package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/jamisonokay/tasty/internal/api"
	"github.com/jamisonokay/tasty/internal/auth"
	"github.com/joho/godotenv"
)

func init() {
    if err := godotenv.Load(); err != nil {
        log.Print("No .env file found")
    }
}

func main() {
    app := fiber.New()
    auth.SetUpAuth(app)
    app.Post("/api/items", api.GetItems)
    app.Listen("0.0.0.0:3005")
}
