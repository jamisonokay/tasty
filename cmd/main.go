package main

import (
    "log"
    "github.com/jamisonokay/tasty/internal/api"
    "github.com/jamisonokay/tasty/internal/auth"
    "github.com/gofiber/fiber/v2"
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
    app.Get("/api/items", api.GetItems)

    app.Listen("0.0.0.0:3005")
}
