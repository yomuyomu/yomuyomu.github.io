package main

import (
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	//"github.com/russross/blackfriday"
	gfm "github.com/shurcooL/github_flavored_markdown"
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

type Reviews map[string]Work

type Work struct {
	Points    int
	Title     string
	Author    string
	Genre     string
	Url       string
	ReviewUrl string
}

func NewUsers() *Users {
	u := Users(make([]string, 0))
	u.update()
	log.Printf("Users count: %d", len(u))
	return &u
}

func NewReviews() *Reviews {
	r := Reviews(make(map[string]Work, 0))
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
	temp := make(map[string]Work, 0)
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
				title := doc2.Find("#workTitle").Find("a").Text()
				link, _ := doc2.Find("#workAuthor-activityName").Find("a").Attr("href")
				author := strings.Split(link, "/")
				genre := doc2.Find("#workGenre").Find("a").Text()
				temp[id] = Work{
					Title:  title,
					Author: author[len(author)-1],
					Genre:  genre,
				}
			}
		})
		if doc.Find(".widget-pagerNext").Length() > 0 {
			page++
		} else {
			break
		}
	}
	// Fliter by overlaps / total <= 1 / 3
	total := len(temp)
	for _, v := range temp {
		overlap := 0
		for _, v2 := range temp {
			if v.Author == v2.Author {
				overlap++
			}
		}
		if overlap*3 > total {
			log.Println("CHEAT USER DETECTED")
			return
		}
	}
	// Apply
	for k, v := range temp {
		(*r)[k] = Work{
			Points:    (*r)[k].Points + 1,
			Title:     v.Title,
			Author:    v.Author,
			Genre:     v.Genre,
			Url:       kakuyomu + "/works/" + k,
			ReviewUrl: kakuyomu + "/works/" + k + "/reviews",
		}
		log.Printf("%s: %d\n", k, (*r)[k].Points)
	}
}

func (r *Reviews) render() {
	log.Println("	Reviews.render")
	pl := (*r).sort()
	//renderJSON(&pl)
	renderMarkdown(&pl)
}

/*func renderJSON(pl *PairList) {
	file, err := os.Create(strconv.FormatInt(time.Now().UnixNano(), 10) + ".json")
	if err != nil {
		log.Fatalln(err)
	}
	enc := json.NewEncoder(file)
	if err := enc.Encode(pl); err != nil {
		log.Println(err)
	}
	log.Println(file.Name())
	file.Close()
}*/

func renderMarkdown(pl *PairList) {
	s := `# カクヨム読者限定＊ランキング

> **これは、読者による読者のためのカクヨムランキング……なのかもしれない** —— @---	

###### 更新時: `
	s += time.Now().Local().String()
	s += `, 有効総数: `
	s += strconv.Itoa(len(*pl))
	s += `

`
	i := 0
	lastPoints := 0
	for _, work := range *pl {
		if lastPoints != work.Value.Points {
			i++
		}
		lastPoints = work.Value.Points
		pre := "# "
		if i == 1 {
			pre = "# "
		} else if i > 1 && i <= 3 {
			pre = "## "
		} else if i > 3 && i <= 10 {
			pre = "### "
		} else {
			pre = "#### "
		}
		s += pre + strconv.Itoa(i) + ": (" + work.Value.Genre + ") [" + work.Value.Title + "](" + work.Value.Url + ") - [＊" + strconv.Itoa(work.Value.Points*100) + "](" + work.Value.ReviewUrl + `)

`
	}
	s += `## 算出方法

[カクヨム検索（暫定版）](http://tueee.net/kakuyomu/users.php?sort=rev&style=0) ⇒ 読者リスト ⇒ 独自不正除去アルゴリズム(※) ⇒ 評価集積 ⇒ ランキング

※ あくまで外部のためアルゴリズムには限界があります`

	// Markdown
	file, err := os.Create("log/" + strconv.FormatInt(time.Now().UnixNano(), 10) + ".md")
	if err != nil {
		log.Fatalln(err)
	}
	_, err = file.WriteString(s)
	if err != nil {
		log.Fatalln(err)
	}
	file.Close()

	file0, err := os.Create("README.md")
	if err != nil {
		log.Fatalln(err)
	}
	_, err = file0.WriteString(s)
	if err != nil {
		log.Fatalln(err)
	}
	file0.Close()

	// HTML
	html := `<html><head><title>カクヨム読者限定＊ランキング</title><meta charset="utf-8"><link href="./gfm.css" media="all" rel="stylesheet" type="text/css" /><link href="//cdnjs.cloudflare.com/ajax/libs/octicons/2.1.2/octicons.css" media="all" rel="stylesheet" type="text/css" /></head><body><article class="markdown-body entry-content" style="padding: 30px;">`
	html += string(gfm.Markdown([]byte(s)))
	html += `</article></body></html>`
	file2, err := os.Create("log/" + strconv.FormatInt(time.Now().UnixNano(), 10) + ".html")
	if err != nil {
		log.Fatalln(err)
	}
	_, err = file2.WriteString(html)
	if err != nil {
		log.Fatalln(err)
	}
	file2.Close()

	file3, err := os.Create("index.html")
	if err != nil {
		log.Fatalln(err)
	}
	_, err = file3.WriteString(html)
	if err != nil {
		log.Fatalln(err)
	}
	file3.Close()
}

func (r *Reviews) sort() PairList {
	pl := make(PairList, len(*r))
	i := 0
	for k, v := range *r {
		pl[i] = Pair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(pl))
	return pl
}

type Pair struct {
	Key   string `json:"work_id"`
	Value Work   `json:"work"`
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value.Points < p[j].Value.Points }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
