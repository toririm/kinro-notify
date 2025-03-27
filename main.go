package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/joho/godotenv"
)

const kinroUrl = "https://kinro.ntv.co.jp/lineup"

type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
}

type Attachment struct {
	Fields   []Field `json:"fields"`
	ImageUrl string  `json:"image_url"`
}

type SlackMessage struct {
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments"`
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Error loading .env file")
	}
	webhookUrl := os.Getenv("SLACK_WEBHOOK")
	c := colly.NewCollector()

	c.OnHTML("section#after_lineup ul", func(e *colly.HTMLElement) {
		message := SlackMessage{}
		e.DOM.Children().Each(func(i int, s *goquery.Selection) {
			img, _ := s.Find("img").Attr("src")
			date := s.Find(".date").Text()
			title := s.Find(".title a").Text()
			fmt.Printf("%s: %s %s\n", date, title, img)
			message.Attachments = append(message.Attachments, Attachment{
				Fields: []Field{
					{Title: title, Value: date},
				},
				ImageUrl: img,
			})
		})
		// Only the latest one
		message.Attachments = message.Attachments[0:1]
		params, _ := json.Marshal(message)
		resp, err := http.PostForm(webhookUrl, url.Values{"payload": {string(params)}})
		if err != nil {
			fmt.Println(err)
		}
		defer resp.Body.Close()
		contents, _ := io.ReadAll(resp.Body)
		fmt.Printf("status: %s, results: %s\n", resp.Status, contents)
	})

	c.Visit(kinroUrl)
}
