package api

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/jamisonokay/tasty/internal/parser"
)

func GetItems(c *fiber.Ctx) error {
    payload := struct {
        Url  string `json:"url"`
    }{}
    if err := c.BodyParser(&payload); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "ok": false,
            "reason": "Can't parse body to string[]",
        })
    }
    urls, nextUrl := parser.GetUrls(payload.Url)
    products, err := parser.Parse(urls)
    if err != nil {
        log.Fatal(err)
    }
    return c.JSON(fiber.Map{
        "products": products,
        "nextUrl": nextUrl,
    })
}
