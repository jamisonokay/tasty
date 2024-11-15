package parser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
	"github.com/jamisonokay/tasty/internal/models"
)

type ResponseBody struct {
    URL string `json:"url"`
}

func GetUrls(url string) ([]string, string) {


    var csrfToken string
    var urls []string
    var response ResponseBody
    c := colly.NewCollector()

    extensions.RandomUserAgent(c)
    extensions.Referer(c)

    c.OnHTML("meta[name='csrf-token']", func(h *colly.HTMLElement) {
        csrfToken = h.Attr("content")
        fmt.Println("Получен CSRF токен:", csrfToken)

        body := bytes.NewBuffer([]byte(`{}`))

        req, err := http.NewRequest("POST", "https://shop.tastycoffee.ru/basket/create_share_link", body)
        if err != nil {
            fmt.Println("Ошибка при создании запроса:", err)
            return
        }

        req.Header.Set("Content-Type", "application/json") 
        req.Header.Set("X-CSRF-TOKEN", csrfToken)         
        req.Header.Set("Referer", url)                   
        req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)") 

        // Передаём куки из colly в запрос
        cookies := c.Cookies("https://shop.tastycoffee.ru")
        for _, cookie := range cookies {
            req.AddCookie(cookie)
        }

        // Отправляем запрос
        client := &http.Client{}
        resp, err := client.Do(req)
        if err != nil {
            fmt.Println("Ошибка при выполнении запроса:", err)
            return
        }
        defer resp.Body.Close()

        respBody, err := io.ReadAll(resp.Body)
        if err != nil {
            fmt.Println("Ошибка при чтении ответа:", err)
            return
        }

        // Разбираем JSON ответ с новым URL
        if err := json.Unmarshal(respBody, &response); err != nil {
            fmt.Println("Ошибка при разборе JSON:", err)
            return
        }
    })

    c.OnHTML("p.goods__name a", func(h *colly.HTMLElement) {
        url := h.Attr("href")
        urls = append(urls, url)
    })

    c.Visit(url)

    return urls, response.URL
}

func Parse(urls []string) ([]models.Product, error) {
    var wg sync.WaitGroup
    productChan := make(chan models.Product, len(urls))
    errChan := make(chan error, len(urls))
    for _, url := range urls {
        wg.Add(1)
        go func(url string) {
            defer wg.Done()
            c := colly.NewCollector()
            product := models.Product{
                Fields: make(map[string]string),
            }
            product.Url = url
            c.OnHTML("h1[itemprop='name']", func(h *colly.HTMLElement) {
                product.Name = strings.TrimSpace(h.Text)
            })
            c.OnHTML("p[itemprop='description']", func(e *colly.HTMLElement) {
                product.Short = strings.TrimSpace(e.Text)
            })
            c.OnHTML("picture img", func(e *colly.HTMLElement) {
                if product.Image == "" {
                    product.Image = e.Attr("src")
                }
            })
            c.OnHTML(".lineCb-wrap", func(e *colly.HTMLElement) {
                e.ForEach(".lineCb-item", func(i int, el *colly.HTMLElement) {
                    styleAttr := el.ChildAttr("span", "style")
                    re := regexp.MustCompile(`width:\s*(\d+)%`)
                    matches := re.FindStringSubmatch(styleAttr)
                    if len(matches) > 1 {
                        value, err := strconv.Atoi(matches[1])
                        if err == nil {
                            text := el.ChildText(".lineCb__text")
                            if text == "Кислотность" && i == 0 {
                                product.Acidity = value
                            } else if text == "Плотность" && i == 1 {
                                product.Density = value
                            }
                        }
                    }
                })
            })

            c.OnHTML("a.priceCb", func(h *colly.HTMLElement) {
                if product.Price == 0 {
                    priceText := strings.TrimSpace(h.Text)
                    priceText = strings.ReplaceAll(priceText, "₽", "")
                    priceText = strings.ReplaceAll(priceText, " ", "")
                    price, err := strconv.Atoi(priceText)
                    if err == nil {
                        product.Price = price
                    }
                }

            })
            c.OnHTML(".descriptionGoods", func(e *colly.HTMLElement) {
                descriptionParts := []string{}
                count := 0
                e.ForEachWithBreak("p", func(i int, h *colly.HTMLElement) bool {
                    if h.DOM.ParentsFiltered(".infoDg").Length() == 0 {
                        clone := h.DOM.Clone()
                        clone.Find("style").Remove()
                        text := strings.TrimSpace(clone.Text())
                        if text != "" {
                            descriptionParts = append(descriptionParts, text)
                            count++
                        }
                    }
                    return count < 4
                })
                product.Description = strings.Join(descriptionParts, "\n\n")
            })
            c.OnHTML(".infoDg .item", func(h *colly.HTMLElement) {
                h.ForEach(".text-p", func(i int, el *colly.HTMLElement) {
                    key := strings.TrimSpace(el.DOM.Contents().Not("span, .tipEl").Text())
                    value := el.ChildText("span.bold-text")
                    if key != "" && value != "" {
                        product.Fields[key] = value
                    }
                })
            })
            if err := c.Visit(url); err != nil {
                errChan <- err
                return
            }
            productChan <- product
        }(url)
    }
    go func() {
        wg.Wait()
        close(productChan)
        close(errChan)
    }()
    var products []models.Product
    var firstError error
    for product := range productChan {
        products = append(products, product)
    }

    for err := range errChan {
        if firstError == nil {
            firstError = err
        }
    }
    return products, firstError
}
