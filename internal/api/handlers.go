package api
import (
    "log"
    "github.com/gofiber/fiber/v2"
    "github.com/jamisonokay/tasty/internal/parser"
    "github.com/jamisonokay/tasty/internal/scraper"
)

func GetItems(c *fiber.Ctx) error {
    urls, err := scraper.GetUrls()
    if err != nil {
        log.Fatal(err)
    }
    products, err := parser.Parse(urls)
    if err != nil {
        log.Fatal(err)
    }
    return c.JSON(products)
}
