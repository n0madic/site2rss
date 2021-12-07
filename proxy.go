package site2rss

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
)

// Document proxy type
type Document = goquery.Document

// Selection proxy type
type Selection = goquery.Selection

// Author proxy type
type Author = feeds.Author

// Enclosure proxy type
type Enclosure = feeds.Enclosure

// Feed proxy type
type Feed = feeds.Feed

// Item proxy type
type Item = feeds.Item

// Link proxy type
type Link = feeds.Link
