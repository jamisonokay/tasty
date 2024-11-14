package auth

import (
    "os"
    "log"

    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/basicauth"
)

func SetUpAuth(app *fiber.App) {
    username, exist := os.LookupEnv("AUTH_USERNAME")
    if !exist {
        log.Fatalln("Can't find AUTH_USERNAME in env")
    }
    password, exist := os.LookupEnv("AUTH_PASSWORD")
    if !exist {
        log.Fatalln("Can't find AUTH_PASSWORD")
    }
    app.Use(basicauth.New(basicauth.Config{
        Users: map[string]string{
            username: password,
        },
        Authorizer: func(user, pass string) bool {
            if user == username && pass == password {
                return true
            }
            return false
        },
        Realm: "Forbidden",
        Unauthorized: func(c *fiber.Ctx) error {
            return c.JSON(fiber.Map{
                "ok": false,
                "reason": "Unauthorized",
            })
        },
    }))
}
