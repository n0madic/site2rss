package site2rss

import (
	"strings"
	"time"
)

// ParseItem is default function for parsing items from remote page
func ParseItem(doc *Document, opts *FindOnPage) *Item {
	item := &Item{
		Link: &Link{Href: doc.Url.String()},
		Id:   doc.Url.String(),
	}
	if opts.Author != "" {
		item.Author = &Author{Name: strings.TrimSpace(doc.Find(opts.Author).First().Text())}
	}
	if opts.Title != "" {
		item.Title = strings.TrimSpace(doc.Find(opts.Title).First().Text())
	}
	if opts.Image != "" {
		imageStr := strings.TrimSpace(doc.Find(opts.Title).First().Text())
		if imageStr != "" {
			item.Enclosure = genEnclosure(imageStr)
		}
	}
	if opts.Date != "" {
		dateStr := strings.TrimSpace(doc.Find(opts.Date).First().Text())
		if dateStr != "" {
			var err error
			item.Created, err = time.Parse(opts.DateFormat, dateStr)
			if err != nil {
				item.Created = HumanTimeParse(dateStr)
			}
		}
	}
	if opts.Description != "" {
		item.Description, _ = doc.Find(opts.Description).Html()
	}
	return item
}

// ParsePage is default function for parsing items from single page
func ParsePage(doc *Document, opts *FindOnPage) *ParseResult {
	return &ParseResult{
		Authors: doc.Find(opts.Author).Map(func(i int, sel *Selection) string {
			return strings.TrimSpace(sel.Text())
		}),
		Dates: doc.Find(opts.Date).Map(func(i int, sel *Selection) string {
			return strings.TrimSpace(sel.Text())
		}),
		Descriptions: doc.Find(opts.Description).Map(func(i int, sel *Selection) string {
			html, _ := sel.Html()
			return html
		}),
		Images: doc.Find(opts.Image).Map(func(i int, sel *Selection) string {
			return sel.AttrOr("src", "")
		}),
		Titles: doc.Find(opts.Title).Map(func(i int, sel *Selection) string {
			return strings.TrimSpace(sel.Text())
		}),
	}
}

// ParseQuery is default function for parsing items from a query by single page
func ParseQuery(sel *Selection, opts *FindOnPage) *Item {
	url := sel.Find(opts.URL).First().AttrOr("href", "")
	if url != "" {
		item := &Item{
			Title: strings.TrimSpace(sel.Find(opts.Title).First().Text()),
			Link:  &Link{Href: url},
			Id:    url,
		}
		if opts.Author != "" {
			item.Author = &Author{Name: strings.TrimSpace(sel.Find(opts.Author).First().Text())}
		}
		if opts.Description != "" {
			item.Description, _ = sel.Find(opts.Description).Html()
		}
		if opts.Image != "" {
			imageStr := sel.Find(opts.Title).First().AttrOr("src", "")
			if imageStr != "" {
				item.Enclosure = genEnclosure(imageStr)
			}
		}
		if opts.Date != "" {
			dateStr := strings.TrimSpace(sel.Find(opts.Date).First().Text())
			if dateStr != "" {
				var err error
				item.Created, err = time.Parse(opts.DateFormat, dateStr)
				if err != nil {
					item.Created = HumanTimeParse(dateStr)
				}
			}
		}
		return item
	}
	return nil
}
