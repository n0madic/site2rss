package site2rss_test

import (
	"testing"

	"github.com/n0madic/site2rss"
)

func TestParseItem(t *testing.T) {
	opts := site2rss.FindOnPage{
		Title:       ".article-title",
		Author:      ".author-name-name",
		Date:        ".author-name-text > div:nth-child(2) > span",
		DateFormat:  "2 January 2006",
		Description: ".article-fulltext",
	}
	rss := site2rss.NewFeed("https://www.sciencealert.com/the-latest", "Science Alert").
		GetLinks("div.titletext > a").
		SetParseOptions(&opts).
		GetItemsFromLinks(site2rss.ParseItem)
	testFeed(t, rss.Feed, &opts)
}

func TestParsePage(t *testing.T) {
	opts := site2rss.FindOnPage{
		Title:       ".titletext",
		Author:      ".category",
		Date:        ".time",
		Image:       ".thumb-article-image > a > img",
		Description: ".introtext-feature",
	}
	rss := site2rss.NewFeed("https://www.sciencealert.com/the-latest", "Science Alert").
		GetLinks(".titletext > a").
		SetParseOptions(&opts).
		GetItemsFromSourcePage(site2rss.ParsePage)
	testFeed(t, rss.Feed, &opts)
}

func TestParseQuery(t *testing.T) {
	opts := site2rss.FindOnPage{
		Title:       ".titletext",
		Date:        ".time",
		Description: ".introtext-feature",
		URL:         ".titletext > a",
	}
	rss := site2rss.NewFeed("https://www.sciencealert.com/the-latest", "Science Alert").
		SetParseOptions(&opts).
		GetItemsFromQuery(".article-item", site2rss.ParseQuery)
	testFeed(t, rss.Feed, &opts)
}

func testFeed(t *testing.T, feed *site2rss.Feed, opts *site2rss.FindOnPage) {
	if len(feed.Items) == 0 {
		t.Error("Expected Feed length is greater than zero")
	}
	for _, item := range feed.Items {
		if opts.Author != "" && item.Author.Name == "" {
			t.Error("Expected non-empty Author Name value")
		}
		if opts.Date != "" && item.Created.IsZero() {
			t.Error("Expected non-empty Date value")
		}
		if opts.Description != "" && item.Description == "" {
			t.Error("Expected non-empty Description value")
		}
		if opts.Title != "" && item.Title == "" {
			t.Error("Expected non-empty Title value")
		}
	}
}
