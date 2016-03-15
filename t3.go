package main

import (
	"encoding/json"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	kakuyomu = `https://kakuyomu.jp`
	tueee    = `http://tueee.net/kakuyomu/users.php?sort=rev&style=0&p=`
)

func main() {
	r := NewReviews()
	r.render()
}

type Users []string

type Reviews map[string]int

type Work struct {
	Points int
	Name   string
}

func NewUsers() *Users {
	u := Users(make([]string, 0))
	u.update()
	log.Printf("Users count: %d", len(u))
	return &u
}

func NewReviews() *Reviews {
	r := Reviews(make(map[string]int, 0))
	r.update()
	log.Printf("Reviews count: %d", len(r))
	return &r
}

func (u *Users) update() {
	log.Println("	Users.update")
	page := 1
	end := false
	for !end {
		doc, err := goquery.NewDocument(tueee + strconv.Itoa(page))
		if err != nil {
			log.Fatalln(err)
		}
		doc.Find("main").Find("article").EachWithBreak(func(i int, s *goquery.Selection) bool {
			var id string
			var points, reviews, follows int
			points, err = strconv.Atoi(strings.TrimPrefix(s.Find("strong").First().Text(), "評価数："))
			if err != nil {
				log.Fatalln(err)
			}
			s.Find("a").Each(func(i2 int, s2 *goquery.Selection) {
				switch i2 {
				case 0:
					link, _ := s2.Attr("href")
					l := strings.Split(link, "/")
					id = l[len(l)-1]
				case 1:
					reviews, err = strconv.Atoi(strings.TrimSuffix(s2.Text(), "件"))
					if err != nil {
						log.Fatalln(err)
					}
				case 2:
					follows, err = strconv.Atoi(strings.TrimSuffix(s2.Text(), "件"))
					if err != nil {
						log.Fatalln(err)
					}
				}
			})
			if reviews < 3 {
				end = true
				return false
			}
			if points <= reviews || points < follows/2 {
				*u = append(*u, id)
				log.Printf("%s: %d %d %d", id, points, reviews, follows)
			}
			return true
		})
		page++
	}
}

func (r *Reviews) update() {
	log.Println("	Reviews.update")
	users := NewUsers()
	for _, u := range *users {
		r.updateByUser(u)
	}
}

func (r *Reviews) updateByUser(userID string) {
	log.Printf("	Reviews.updateByUser(%s)", userID)
	page := 1
	idAuth := make(map[string]string, 0)
	for {
		doc, err := goquery.NewDocument(kakuyomu + "/users/" + userID + "/reviews?page=" + strconv.Itoa(page))
		if err != nil {
			log.Fatalln(err)
		}
		doc.Find("#reviews-list").First().Find("[itemprop=\"reviewBody\"]").Each(func(i int, s *goquery.Selection) {
			link, _ := s.Attr("href")
			l := strings.Split(link, "/")
			if id := l[len(l)-1]; id != "" {
				doc2, err := goquery.NewDocument(kakuyomu + "/works/" + id)
				if err != nil {
					log.Fatalln(err)
				}
				link, _ := doc2.Find("#workAuthor-activityName").Find("a").Attr("href")
				l := strings.Split(link, "/")
				idAuth[id] = l[len(l)-1]
			}
		})
		if doc.Find(".widget-pagerNext").Length() > 0 {
			page++
		} else {
			break
		}
	}
	// Fliter by overlaps / total <= 1 / 3
	total := len(idAuth)
	for _, v := range idAuth {
		overlap := 0
		for _, v2 := range idAuth {
			if v == v2 {
				overlap++
			}
		}
		if overlap*3 > total {
			log.Println("CHEAT USER DETECTED")
			return
		}
	}
	// Apply
	for k, _ := range idAuth {
		(*r)[k] = (*r)[k] + 1
		log.Printf("%s: %d\n", k, (*r)[k])
	}
}

func (r *Reviews) render() {
	log.Println("	Reviews.render")
	pl := (*r).sort()
	/*for i, p := range pl {
		fmt.Printf("%d: %s %d\n", i+1, p.Key, p.Value)
	}*/
	log.Println("	Reviews.renderJSON")
	file, err := os.Create(strconv.FormatInt(time.Now().UnixNano(), 10) + ".json")
	if err != nil {
		log.Fatalln(err)
	}
	enc := json.NewEncoder(file)
	if err := enc.Encode(&pl); err != nil {
		log.Println(err)
	}
	log.Println(file.Name())
	file.Close()
}

func (r Reviews) sort() PairList {
	pl := make(PairList, len(map[string]int(r)))
	i := 0
	for k, v := range r {
		pl[i] = Pair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(pl))
	return pl
}

type Pair struct {
	Key   string `json:"work_id"`
	Value int    `json:"count"`
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
