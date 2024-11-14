package parser
import (
    "github.com/gocolly/colly"
    "regexp"
    "strconv"
    "strings"
    "sync"
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
