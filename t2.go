package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	kakuyomu = `https://kakuyomu.jp`
	tueee    = `http://tueee.net/kakuyomu/?keyword=&sort=pnt&contest=1&p=`
)

type Work struct {
	Title     string
	ID        string
	Star      int
	Txt       int
	TxtAve    int
	Author    *User
	Reviews   []*Review
	Followers []*User
}

type User struct {
	Name string
	ID   string
	FW   int
	FU   int
	W    int
	N    int
	R    int
}

type Review struct {
	PointOnly bool
	Reviewer  *User
}

func worksByTueee(count int) []*Work {
	var works []*Work
	for i := 0; i < count; i++ {
		doc, err := goquery.NewDocument(tueee + strconv.Itoa(i+1))
		if err != nil {
			log.Fatalln(err)
		}
		doc.Find("main").Find("article").Each(func(i int, s *goquery.Selection) {
			work := &Work{
				ID:     strings.TrimPrefix(s.AttrOr("id", ""), "n"),
				Star:   func() int { s, _ := strconv.Atoi(s.AttrOr("data-pnt", "")); return s }(),
				Txt:    func() int { s, _ := strconv.Atoi(s.AttrOr("data-txt", "")); return s }(),
				TxtAve: func() int { s, _ := strconv.Atoi(s.AttrOr("data-ave", "")); return s }(),
			}
			work.update()
			works = append(works, work)
		})
	}
	return works
}

func (w *Work) update() {
	doc, err := goquery.NewDocument(kakuyomu + "/works/" + w.ID)
	if err != nil {
		log.Fatalln(err)
	}
	w.Title = doc.Find("#workTitle").First().Find("a").Text()
	link, _ := doc.Find("#workAuthor-activityName").First().Find("a").First().Attr("href")
	s := strings.Split(link, "/")
	u := &User{
		ID: s[len(s)-1],
	}
	u.update()
	w.Author = u
	w.reviews()
}

func (w *Work) reviews() {
	w.Reviews = []*Review{}
	page := 1
	for {
		doc, err := goquery.NewDocument(kakuyomu + "/works/" + w.ID + "/reviews?page=" + strconv.Itoa(page))
		if err != nil {
			log.Fatalln(err)
		}
		doc.Find("#workReview-list").First().Find("article").Each(func(i int, s *goquery.Selection) {
			r := &Review{}
			if s.HasClass("isOnlyPoints") {
				link, _ := s.Find("a").Attr("href")
				l := strings.Split(link, "/")
				r.Reviewer = &User{
					ID: l[len(l)-1],
				}
				r.PointOnly = true
			} else {
				link, _ := s.Find(".workReview-reviewTitleAuthor").Attr("href")
				l := strings.Split(link, "/")
				r.Reviewer = &User{
					ID: l[len(l)-1],
				}
				r.PointOnly = false
			}
			if r != nil {
				w.Reviews = append(w.Reviews, r)
			}
		})
		next := doc.Find(".widget-pagerNext").Length() > 0
		if !next {
			break
		} else {
			page++
		}
	}
}

func (u *User) update() {
	doc, err := goquery.NewDocument(kakuyomu + "/users/" + u.ID)
	if err != nil {
		log.Fatalln(err)
	}
	header := doc.Find("#widget-user-header").First()
	u.Name = header.Find("#user-name-activityName").First().Find("a").Text()
	header.Find("#user-meta").First().Find(".user-following-count").Each(func(i int, s *goquery.Selection) {
		if i > 1 {
			return
		}
		count, err := strconv.Atoi(strings.Replace(s.Text(), ",", "", 1))
		if err != nil {
			log.Fatalln(err)
		}
		switch i {
		case 0:
			u.FW = count
		case 1:
			u.FU = count
		}
	})
	header.Find("#user-nav").First().Find("ul").First().Find("li").Each(func(i int, s *goquery.Selection) {
		if i < 1 || i > 3 {
			return
		}
		count, err := strconv.Atoi(s.Find(".widget-user-navCount").First().Text())
		if err != nil {
			log.Fatalln(err)
		}
		switch i {
		case 1:
			u.W = count
		case 2:
			u.N = count
		case 3:
			u.R = count
		}
	})
}

func renderJSON(works []*Work) {
	file, err := os.Create("3000-" + strconv.FormatInt(time.Now().UnixNano(), 10) + ".json")
	if err != nil {
		log.Fatalln(err)
	}
	enc := json.NewEncoder(file)
	if err := enc.Encode(&works); err != nil {
		log.Println(err)
	}
	file.Close()
}

func render(works []*Work) {
	file, err := os.Create("3000-" + strconv.FormatInt(time.Now().UnixNano(), 10) + ".txt")
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(file.Name())
	for i, w := range works {
		cross := []int{}
		for _, r := range w.Reviews {
			for i3, w2 := range works {
				if r.Reviewer.ID == w2.Author.ID {
					cross = append(cross, i3+1)
				}
			}
		}
		file.WriteString(
			strconv.Itoa(i+1) + ": " + w.Title + " - ★:" + strconv.Itoa(w.Star) + " T:" + strconv.Itoa(w.Txt) + "(" + strconv.Itoa(w.TxtAve) + ") " + "\n" +
				"　" + w.Author.Name + " - FW:" + strconv.Itoa(w.Author.FW) + " FU:" + strconv.Itoa(w.Author.FU) + " W:" + strconv.Itoa(w.Author.W) + " N:" + strconv.Itoa(w.Author.N) + " R:" + strconv.Itoa(w.Author.R) + "\n" +
				fmt.Sprintf("　⇔:%v\n", cross),
		)
	}
	file.Close()
}

func main() {
	w := worksByTueee(100)
	render(w)
	renderJSON(w)
	//userMeta(userByWork("4852201425154980055"))
}
