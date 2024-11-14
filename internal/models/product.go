package models

type Product struct {
    Name        string            `json:"name"`
    Image       string            `json:"image"`
    Short       string            `json:"short"`
    Price       int               `json:"price"`
    Acidity     int               `json:"acidity"`
    Density     int               `json:"density"`
    Description string            `json:"description"`
    Fields      map[string]string `json:"fields"`
    Url         string            `json:"url"`
}
