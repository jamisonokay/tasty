package scraper

import (
	"errors"
	"os"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/launcher"
)

func GetUrls() ([]string, error) {
    path, _ := launcher.LookPath()
    u := launcher.New().Bin(path).MustLaunch()
    browser := rod.New().ControlURL(u)
    defer browser.MustClose()
    page := browser.MustPage("https://shop.tastycoffee.ru/login")

    email, e := os.LookupEnv("EMAIL")
    if !e {
        return nil, errors.New("Can't find email in env")
    }
    password, p := os.LookupEnv("PASSWORD")
    if !p {
        return nil, errors.New("Can't find password in env")
    }

    page.MustElement("#login_email").MustInput(email)
    page.MustElement("#login_password").MustInput(password).MustType(input.Enter)
    page.MustWaitLoad()
    page.MustNavigate("https://shop.tastycoffee.ru/basket")

    var links []string
    elements := page.MustElements(".goods-item")
    for _, element := range elements {
        linkElement := element.MustElement(".goods__name a")
        link := linkElement.MustProperty("href").String()
        links = append(links, link)
    }

    return links, nil
}
