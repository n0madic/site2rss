# Site2RSS golang library

Go library for scraping the site and creating RSS feeds.

## Usage

```go
package main

import (
    "net/http"
    "time"

    "github.com/n0madic/site2rss"
)

func rssRequest(w http.ResponseWriter, r *http.Request) {
    rss, err := site2rss.NewFeed("https://www.sciencealert.com/the-latest", "Science Alert").
        GetLinks("div.titletext > a").
        GetFeedItems(func(doc *site2rss.Document) *site2rss.Item {
            author := doc.Find(".author-name-name").First().Text()
            title := doc.Find(".article-title").First().Text()
            created, _ := time.Parse("02 Jan 2006", doc.Find(".author-name-date").First().Text())
            desc, _ := doc.Find(".article-fulltext").Html()
            return &site2rss.Item{
                Title:       title,
                Author:      &site2rss.Author{Name: author},
                Link:        &site2rss.Link{Href: doc.Url.String()},
                Id:          doc.Url.String(),
                Description: desc,
                Created:     created,
            }
        }).GetRSS()
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte(err.Error()))
    } else {
        w.Header().Set("Content-Type", "application/xml")
        w.Write([]byte(rss))
    }
}

func main() {
    http.HandleFunc("/", rssRequest)
    http.ListenAndServe(":3000", nil)
}
```