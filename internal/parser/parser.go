package parser

import (
	"strconv"
	"strings"
	"sync"

	"github.com/gocolly/colly"
	"github.com/jamisonokay/tasty/internal/models"
)


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
            c.OnHTML("meta[itemprop='description']", func(e *colly.HTMLElement) {
                product.Short = strings.TrimSpace(e.Attr("content"))
            })
            c.OnHTML("meta[itemprop='image']", func(e *colly.HTMLElement) {
                if product.Image == "" {
                    product.Image = e.Attr("content")
                }
            })
            c.OnHTML(".tc-tile__scale", func(e *colly.HTMLElement) {
                e.ForEach("div > div", func(i int, el *colly.HTMLElement) {
                    t := el.Text
                    w := el.ChildAttr("object.tc-progress__object", "width")
                    cleanW := strings.ReplaceAll(w, "%", "")
                    value, err := strconv.Atoi(cleanW)
                    if err != nil {
                        value = 0
                    }
                    if t == "Плотность" && product.Density == 0 {
                        product.Density = value
                    } else if t == "Кислотность" && product.Acidity == 0 {
                        product.Acidity = value
                    }

                })
            })


            c.OnHTML("span.text-nowrap", func(h *colly.HTMLElement) {
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
            c.OnHTML(".content-static", func(e *colly.HTMLElement) {
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
            c.OnHTML(".additional-info ul > li", func(h *colly.HTMLElement) {
                key := h.DOM.Find("div p.inline").Text()
                value := h.DOM.Find("p:last-child").Text()
                key = strings.TrimSpace(key)
                value = strings.TrimSpace(value)

                if key != "" && value != "" {
                    product.Fields[key] = value
                }
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
