package site2rss

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"path"
	"time"

	"github.com/PuerkitoBio/goquery"
	timeparse "github.com/dsoprea/go-time-parse"
	"github.com/gorilla/feeds"
	"golang.org/x/net/html/charset"
)

func genEnclosure(image string) *feeds.Enclosure {
	return &feeds.Enclosure{
		Length: "-1",
		Type:   mime.TypeByExtension(path.Ext(image)),
		Url:    image,
	}
}

func getNewDocumentFromURL(url string) (*goquery.Document, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}
	return goquery.NewDocumentFromReader(res.Body)
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

// ConvertToUTF8 string from any encoding
func ConvertToUTF8(str string, origEncoding string) string {
	strBytes := []byte(str)
	byteReader := bytes.NewReader(strBytes)
	reader, _ := charset.NewReaderLabel(origEncoding, byteReader)
	strBytes, _ = ioutil.ReadAll(reader)
	return string(strBytes)
}
