package site2rss

import (
	"fmt"
	"net/url"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
)

// Site2RSS object
type Site2RSS struct {
	baseURL      string
	feed         *feeds.Feed
	Links        []string
	MaxFeedItems int
	SourceURL    *url.URL
	wg           sync.WaitGroup
}

type itemCallback func(doc *goquery.Document) *feeds.Item

// NewFeed return a new Site2RSS feed object
func NewFeed(source string, title string) *Site2RSS {
	s := &Site2RSS{}
	s.MaxFeedItems = 10
	sourceURL, err := url.Parse(source)
	if err != nil {
		panic("invalid url")
	}
	s.baseURL = fmt.Sprintf("%s://%s", sourceURL.Scheme, sourceURL.Hostname())
	s.SourceURL = sourceURL
	s.feed = &feeds.Feed{
		Title: title,
		Link:  &feeds.Link{Href: s.baseURL},
	}
	return s
}

// AbsoluteURL makes the relative URL absolute
func (s *Site2RSS) AbsoluteURL(rpath string) string {
	abspath := rpath
	u, err := url.Parse(rpath)
	if err == nil {
		abspath = s.SourceURL.ResolveReference(u).String()
	}
	return abspath
}

// MakeAllLinksAbsolute makes all links absolute in document
func (s *Site2RSS) MakeAllLinksAbsolute(doc *goquery.Document) {
	doc.Find("a,img").Each(func(i int, sel *goquery.Selection) {
		if link, ok := sel.Attr("src"); link != "" && ok {
			u, _ := url.Parse(link)
			if !u.IsAbs() {
				sel.SetAttr("src", s.AbsoluteURL(link))
			}
		}
		if link, ok := sel.Attr("href"); link != "" && ok {
			u, _ := url.Parse(link)
			if !u.IsAbs() {
				sel.SetAttr("href", s.AbsoluteURL(link))
			}
		}
	})
}

// GetLinks get a list of links by pattern
func (s *Site2RSS) GetLinks(linkPattern string) *Site2RSS {
	doc, err := goquery.NewDocument(s.SourceURL.String())
	if err == nil {
		links := doc.Find(linkPattern).Map(func(i int, s *goquery.Selection) string {
			link, _ := s.Attr("href")
			return link
		})
		chunk := s.MaxFeedItems
		if len(links) < s.MaxFeedItems {
			chunk = len(links)
		}
		s.Links = links[:chunk]
	}
	return s
}

// GetFeedItems extracts details using a user-defined function
func (s *Site2RSS) GetFeedItems(f itemCallback) *Site2RSS {
	feedItems := make([]*feeds.Item, len(s.Links))
	for i := 0; i < len(s.Links); i++ {
		url := s.AbsoluteURL(s.Links[i])
		s.wg.Add(1)
		go func(url string, item **feeds.Item) {
			defer s.wg.Done()
			itemDoc, err := goquery.NewDocument(url)
			if err == nil {
				s.MakeAllLinksAbsolute(itemDoc)
				*item = f(itemDoc)
			}
		}(url, &feedItems[i])
	}
	s.wg.Wait()
	s.feed.Items = feedItems
	return s
}

// GetRSS return feed xml
func (s *Site2RSS) GetRSS() (string, error) {
	return s.feed.ToRss()
}
