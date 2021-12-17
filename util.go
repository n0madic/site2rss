package site2rss

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	timeparse "github.com/dsoprea/go-time-parse"
	"github.com/goodsign/monday"
	"github.com/gorilla/feeds"
	"golang.org/x/net/html/charset"
)

func checkYear(date *time.Time) time.Time {
	year, month, day := date.Date()
	if year == 0 {
		now := time.Now()
		year = now.Year()
		if month == time.January && day == 1 {
			month = now.Month()
			day = now.Day()
		}
		return time.Date(year, month, day, date.Hour(), date.Minute(), date.Second(), date.Nanosecond(), date.Location())
	}
	return *date
}

func genEnclosure(image string) *feeds.Enclosure {
	return &feeds.Enclosure{
		Length: "-1",
		Type:   mime.TypeByExtension(path.Ext(image)),
		Url:    image,
	}
}

func getNewDocumentFromURL(sourceURL string) (*goquery.Document, error) {
	res, err := http.Get(sourceURL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	doc.Url, err = url.Parse(sourceURL)
	if err != nil {
		return nil, fmt.Errorf("invalid url")
	}

	return doc, nil
}

// HumanTimeParse from string
func HumanTimeParse(d string) time.Time {
	defer func() {
		recover()
	}()
	duration, phraseType, err := timeparse.ParseDuration(d)
	if err == nil && phraseType == timeparse.PhraseTypeTime {
		return time.Now().Add(duration)
	}
	return time.Time{}
}

// TimeParse from string
func TimeParse(layout, dateStr string) time.Time {
	if layout != "" {
		timeLocaleDetector := monday.NewLocaleDetector()
		parsedDate, err := timeLocaleDetector.Parse(layout, dateStr)
		if err == nil {
			return checkYear(&parsedDate)
		}
		parsedDate, err = time.Parse(layout, dateStr)
		if err == nil {
			return checkYear(&parsedDate)
		}
	}
	return HumanTimeParse(dateStr)
}

// ConvertToUTF8 string from any encoding
func ConvertToUTF8(str string, origEncoding string) string {
	strBytes := []byte(str)
	byteReader := bytes.NewReader(strBytes)
	reader, _ := charset.NewReaderLabel(origEncoding, byteReader)
	strBytes, _ = ioutil.ReadAll(reader)
	return string(strBytes)
}

func stringIsFiltered(str string, filters []string) bool {
	for _, filter := range filters {
		if strings.Contains(strings.ToLower(str), strings.ToLower(filter)) {
			return true
		}
	}
	return false
}
