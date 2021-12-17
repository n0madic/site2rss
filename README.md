# Site2RSS golang library

Go library for scraping the site and creating RSS feeds.

## Usage

### Parse feed items from remote pages

```go
package main

import (
    "net/http"

    "github.com/n0madic/site2rss"
)

func rssRequest(w http.ResponseWriter, r *http.Request) {
    rss, err := site2rss.NewFeed("https://www.sciencealert.com/the-latest", "Science Alert").
        GetLinks("div.titletext > a").
        SetParseOptions(&site2rss.FindOnPage{
            Title:       ".article-title",
            Author:      ".author-name-name",
            Date:        ".author-name-date",
            DateFormat:  "02 Jan 2006",
            Description: ".article-fulltext",
        }).
        GetItemsFromLinks(site2rss.ParseItem).
        GetRSS()
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

### Parse remote pages with user-defined function

```go
package main

import (
    "net/http"
    "strings"
    "time"

    "github.com/n0madic/site2rss"
)

func rssRequest(w http.ResponseWriter, r *http.Request) {
    rss, err := site2rss.NewFeed("https://www.sciencealert.com/the-latest", "Science Alert").
        GetLinks("div.titletext > a").
        GetItemsFromLinks(func(doc *site2rss.Document, opts *site2rss.FindOnPage) *site2rss.Item {
            author := doc.Find(".author-name-name").First().Text()
            title := doc.Find(".article-title").First().Text()
            created, _ := time.Parse("02 Jan 2006", strings.TrimSpace(doc.Find(".author-name-date").First().Text()))
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

### Parse feed items from source page

```go
package main

import (
    "net/http"

    "github.com/n0madic/site2rss"
)

func rssRequest(w http.ResponseWriter, r *http.Request) {
    rss, err := site2rss.NewFeed("https://www.sciencealert.com/the-latest", "Science Alert").
        GetLinks(".titletext > a").
        SetParseOptions(&site2rss.FindOnPage{
            Title:       ".titletext",
            Author:      ".category",
            Date:        ".time",
            Image:       ".thumb-article-image > a > img",
            Description: ".introtext-feature",
        }).
        GetItemsFromSourcePage(site2rss.ParsePage).
        GetAtom()
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

### Parse feed items from a query by source page

```go
package main

import (
    "net/http"

    "github.com/n0madic/site2rss"
)

func rssRequest(w http.ResponseWriter, r *http.Request) {
    rss, err := site2rss.NewFeed("https://www.sciencealert.com/the-latest", "Science Alert").
        SetParseOptions(&site2rss.FindOnPage{
            Title:       ".titletext",
            Date:        ".time",
            Description: ".introtext-feature",
            URL:         ".titletext > a",
        }).
        GetItemsFromQuery(".article-item", site2rss.ParseQuery).
        GetAtom()
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

#### Or with user-defined function:

```go
package main

import (
    "net/http"

    "github.com/n0madic/site2rss"
)

func rssRequest(w http.ResponseWriter, r *http.Request) {
    rss, err := site2rss.NewFeed("https://www.sciencealert.com/the-latest", "Science Alert").
        GetItemsFromQuery(".article-item",
            func(doc *site2rss.Selection, opts *site2rss.FindOnPage) *site2rss.Item {
                url := "https://www.sciencealert.com" +
                    doc.Find(".titletext > a").First().AttrOr("href", "")
                desc, _ := doc.Find(".introtext-feature").Html()
                return &site2rss.Item{
                    Title:       doc.Find(".titletext").First().Text(),
                    Link:        &site2rss.Link{Href: url},
                    Id:          url,
                    Description: desc,
                    Created:     site2rss.HumanTimeParse(doc.Find(".time").First().Text()),
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

### Filter items

You can filter feed items by titles, text in description, text blocks, or CSS selector

```go
site2rss.FilterItems(site2rss.Filters{
    Description: []string{"spam"},
    Selector: []string{".comments"},
    Text: []string{"See also:"},
    Title: []string{"advertising"},
})
```
