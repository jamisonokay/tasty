package scraper

import (
	"errors"
	"log"
	"os"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/launcher"
)

func GetUrls() ([]string, error) {
    rodRemote, exist:= os.LookupEnv("ROD_REMOTE")
    if !exist {
        log.Fatal("Cant't connect to browser havent remote")
        return nil, errors.New("ROD_REMOTE env is not set")
    }
    launch := launcher.MustNewManaged(rodRemote)
    browser := rod.New().Client(launch.MustClient()).MustConnect()
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
    defer browser.MustClose()
    return links, nil
}
