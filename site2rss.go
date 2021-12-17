package site2rss

import (
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
)

// Filters for item cleaning
type Filters struct {
	// Skip item with the following words in the description
	Descriptions []string
	// Remove the following selectors from content
	Selectors []string
	// Remove blocks of text that contain the following words
	Text []string
	// Skip items with the following words in the title
	Titles []string
}

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
	Feed         *Feed
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
	s.Feed = &Feed{
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
			u, err := url.Parse(link)
			if err == nil && !u.IsAbs() {
				sel.SetAttr("src", s.AbsoluteURL(link))
			}
		}
		if link, ok := sel.Attr("href"); link != "" && ok {
			u, err := url.Parse(link)
			if err == nil && !u.IsAbs() {
				sel.SetAttr("href", s.AbsoluteURL(link))
			}
		}
	})
}

// GetLinks get a list of links by pattern
func (s *Site2RSS) GetLinks(linkPattern string) *Site2RSS {
	var err error
	s.sourceDoc, err = getNewDocumentFromURL(s.SourceURL.String())
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
	s.Feed.Items = make([]*feeds.Item, len(s.Links))
	for i := 0; i < len(s.Links); i++ {
		s.wg.Add(1)
		go func(url string, item **feeds.Item) {
			defer s.wg.Done()
			itemDoc, err := getNewDocumentFromURL(url)
			if err == nil {
				s.MakeAllLinksAbsolute(itemDoc)
				*item = f(itemDoc, s.parseOpts)
			}
		}(s.Links[i], &s.Feed.Items[i])
	}
	s.wg.Wait()
	return s
}

// GetItemsFromQuery extracts feed items from a query by source page
func (s *Site2RSS) GetItemsFromQuery(docPattern string, f queryCallback) *Site2RSS {
	var err error
	s.sourceDoc, err = getNewDocumentFromURL(s.SourceURL.String())
	if err == nil {
		s.sourceDoc.Find(docPattern).Each(func(i int, sel *goquery.Selection) {
			item := f(sel, s.parseOpts)
			if len(s.Feed.Items) < s.MaxFeedItems && item != nil {
				item.Link.Href = s.AbsoluteURL(item.Link.Href)
				s.Feed.Items = append(s.Feed.Items, item)
			}
		})
	}
	return s
}

// GetItemsFromSourcePage extracts feed items from source page
func (s *Site2RSS) GetItemsFromSourcePage(f pageCallback) *Site2RSS {
	if len(s.Links) > 0 {
		s.Feed.Items = make([]*feeds.Item, len(s.Links))
		for i := 0; i < len(s.Links); i++ {
			s.Feed.Items[i] = &feeds.Item{
				Id:   s.Links[i],
				Link: &feeds.Link{Href: s.Links[i]},
			}
			parse := f(s.sourceDoc, s.parseOpts)
			if len(parse.Authors) >= len(s.Links) && parse.Authors[i] != "" {
				s.Feed.Items[i].Author = &feeds.Author{Name: parse.Authors[i]}
			}
			if len(parse.Descriptions) >= len(s.Links) {
				s.Feed.Items[i].Description = parse.Descriptions[i]
			}
			if len(parse.Dates) >= len(s.Links) && parse.Dates[i] != "" {
				s.Feed.Items[i].Created = TimeParse(s.parseOpts.DateFormat, parse.Dates[i])
			}
			if len(parse.Images) >= len(s.Links) && parse.Images[i] != "" {
				s.Feed.Items[i].Enclosure = genEnclosure(s.AbsoluteURL(parse.Images[i]))
			}
			if len(parse.Titles) >= len(s.Links) {
				s.Feed.Items[i].Title = parse.Titles[i]
			}
		}
	}
	return s
}

// FilterItems for clean items
func (s *Site2RSS) FilterItems(filters Filters) *Site2RSS {
	var items []*feeds.Item

	for _, item := range s.Feed.Items {
		if stringIsFiltered(item.Title, filters.Titles) ||
			stringIsFiltered(item.Description, filters.Descriptions) {
			continue
		}

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(item.Description))
		if err == nil {
			doc.Find("script").Remove()

			if len(filters.Selectors) > 0 {
				doc.Find(strings.Join(filters.Selectors, ", ")).Remove()
			}

			if len(filters.Text) > 0 {
				var searchText []string
				for _, text := range filters.Text {
					searchText = append(searchText, fmt.Sprintf("p:contains('%s')", text))
					searchText = append(searchText, fmt.Sprintf("div:contains('%s')", text))
				}
				doc.Find(strings.Join(searchText, ", ")).Remove()
			}

			filteredContent, err := doc.Html()
			if err == nil {
				item.Description = filteredContent
			}
		}
		items = append(items, item)
	}

	s.Feed.Items = items

	return s
}

// GetAtom return feed xml
func (s *Site2RSS) GetAtom() (string, error) {
	return s.Feed.ToAtom()
}

// GetRSS return feed xml
func (s *Site2RSS) GetRSS() (string, error) {
	return s.Feed.ToRss()
}
