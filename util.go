package site2rss

import (
	"mime"
	"path"
	"time"

	timeparse "github.com/dsoprea/go-time-parse"
	"github.com/gorilla/feeds"
)

func humanTimeParse(d string) time.Time {
	defer func() {
		recover()
	}()
	duration, phraseType, err := timeparse.ParseDuration(d)
	if err == nil && phraseType == timeparse.PhraseTypeTime {
		return time.Now().Add(duration)
	}
	return time.Time{}
}

func genEnclosure(image string) *feeds.Enclosure {
	return &feeds.Enclosure{
		Length: "-1",
		Type:   mime.TypeByExtension(path.Ext(image)),
		Url:    image,
	}
}
