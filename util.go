package site2rss

import (
	"bytes"
	"io/ioutil"
	"mime"
	"path"
	"time"

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
