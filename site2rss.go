package site2rss

import (
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
)

// FindOnPage settings for parse page to feed item
type FindOnPage struct {
	Author      string
	Date        string
	DateFormat  string
	Description string
	Image       string
	Title       string
	URL         string
}

// Site2RSS object
type Site2RSS struct {
	baseURL      string
	feed         *feeds.Feed
	Links        []string
	MaxFeedItems int
	parseOpts    *FindOnPage
	sourceDoc    *Document
	SourceURL    *url.URL
	wg           sync.WaitGroup
}

// ParseResult return results of parsing single page
type ParseResult struct {
	Authors      []string
	Dates        []string
	Descriptions []string
	Images       []string
	Titles       []string
}

type itemCallback func(doc *Document, opts *FindOnPage) *Item
type pageCallback func(doc *Document, opts *FindOnPage) *ParseResult
type queryCallback func(s *Selection, opts *FindOnPage) *Item

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

// SetMaxFeedItems set max feed items
func (s *Site2RSS) SetMaxFeedItems(max int) *Site2RSS {
	s.MaxFeedItems = max
	return s
}

// SetParseOptions for parse page
func (s *Site2RSS) SetParseOptions(opts *FindOnPage) *Site2RSS {
	s.parseOpts = opts
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
func (s *Site2RSS) MakeAllLinksAbsolute(doc *Document) {
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
	var err error
	s.sourceDoc, err = goquery.NewDocument(s.SourceURL.String())
	if err == nil {
		links := s.sourceDoc.Find(linkPattern).Map(func(i int, sel *goquery.Selection) string {
			link, _ := sel.Attr("href")
			return s.AbsoluteURL(link)
		})
		chunk := s.MaxFeedItems
		if len(links) < s.MaxFeedItems {
			chunk = len(links)
		}
		s.Links = append([]string(nil), links[:chunk]...)
	}
	return s
}

// GetItemsFromLinks extracts details from remote links using a user-defined function
func (s *Site2RSS) GetItemsFromLinks(f itemCallback) *Site2RSS {
	s.feed.Items = make([]*feeds.Item, len(s.Links))
	for i := 0; i < len(s.Links); i++ {
		s.wg.Add(1)
		go func(url string, item **feeds.Item) {
			defer s.wg.Done()
			itemDoc, err := goquery.NewDocument(url)
			if err == nil {
				s.MakeAllLinksAbsolute(itemDoc)
				*item = f(itemDoc, s.parseOpts)
			}
		}(s.Links[i], &s.feed.Items[i])
	}
	s.wg.Wait()
	return s
}

// GetItemsFromQuery extracts feed items from a query by source page
func (s *Site2RSS) GetItemsFromQuery(docPattern string, f queryCallback) *Site2RSS {
	var err error
	s.sourceDoc, err = goquery.NewDocument(s.SourceURL.String())
	if err == nil {
		s.sourceDoc.Find(docPattern).Each(func(i int, sel *goquery.Selection) {
			item := f(sel, s.parseOpts)
			if len(s.feed.Items) < s.MaxFeedItems && item != nil {
				item.Link.Href = s.AbsoluteURL(item.Link.Href)
				s.feed.Items = append(s.feed.Items, item)
			}
		})
	}
	return s
}

// GetItemsFromSourcePage extracts feed items from source page
func (s *Site2RSS) GetItemsFromSourcePage(f pageCallback) *Site2RSS {
	if len(s.Links) > 0 {
		s.feed.Items = make([]*feeds.Item, len(s.Links))
		for i := 0; i < len(s.Links); i++ {
			s.feed.Items[i] = &feeds.Item{
				Id:   s.Links[i],
				Link: &feeds.Link{Href: s.Links[i]},
			}
			parse := f(s.sourceDoc, s.parseOpts)
			if len(parse.Authors) >= len(s.Links) && parse.Authors[i] != "" {
				s.feed.Items[i].Author = &feeds.Author{Name: parse.Authors[i]}
			}
			if len(parse.Descriptions) >= len(s.Links) {
				s.feed.Items[i].Description = parse.Descriptions[i]
			}
			if len(parse.Dates) >= len(s.Links) && parse.Dates[i] != "" {
				created, err := time.Parse(s.parseOpts.DateFormat, parse.Dates[i])
				if err == nil {
					s.feed.Items[i].Created = created
				} else {
					s.feed.Items[i].Created = HumanTimeParse(parse.Dates[i])
				}
			}
			if len(parse.Images) >= len(s.Links) && parse.Images[i] != "" {
				s.feed.Items[i].Enclosure = genEnclosure(s.AbsoluteURL(parse.Images[i]))
			}
			if len(parse.Titles) >= len(s.Links) {
				s.feed.Items[i].Title = parse.Titles[i]
			}
		}
	}
	return s
}

// GetAtom return feed xml
func (s *Site2RSS) GetAtom() (string, error) {
	return s.feed.ToAtom()
}

// GetRSS return feed xml
func (s *Site2RSS) GetRSS() (string, error) {
	return s.feed.ToRss()
}
