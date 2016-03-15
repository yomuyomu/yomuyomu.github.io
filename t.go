package main

import (
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	baseUrl = `https://kakuyomu.jp`
)

var users = []string{
	"---",
}

func reviews(user string) {
	doc, err := goquery.NewDocument(baseUrl + "/users/" + user + "/reviews")
	if err != nil {
		log.Fatal(err)
	}
	doc.Find("[itemtype=\"https://schema.org/CreativeWork\"]").Each(func(i int, s *goquery.Selection) {
		rb := s.Find("[itemprop=\"reviewBody\"]")
		text := rb.Text()
		//link, _ := rb.Attr("href")
		//link = baseUrl + link
		author := s.Find("[itemprop=\"author\"]").Last().Text()
		name := s.Find("[itemprop=\"name\"]").Text()
		genre := s.Find("[itemprop=\"genre\"]").Text()
		star := s.Find(".widget-work-reviewPoints").Text()
		star = strings.TrimPrefix(star, "★")
		log.Printf("Review %d: %s %s %s %s %s", i, text, name, genre, author, star)
	})
}

func main() {
	for _, user := range users {
		reviews(user)
	}
}
