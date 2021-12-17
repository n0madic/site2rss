package site2rss_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/n0madic/site2rss"
)

func TestTimeParse(t *testing.T) {
	now := time.Now()
	type args struct {
		layout  string
		dateStr string
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{
			name: "Full layout",
			args: args{
				layout:  "Mon, 2 Jan 2006 15:04:05",
				dateStr: "Fri, 17 Dec 2021 11:41:59",
			},
			want: time.Date(2021, 12, 17, 11, 41, 59, 0, time.UTC),
		},
		{
			name: "Russian layout",
			args: args{
				layout:  "Monday, 2 January 2006 15:04:05",
				dateStr: "Пятница, 17 декабря 2021 11:41:59",
			},
			want: time.Date(2021, 12, 17, 11, 41, 59, 0, time.UTC),
		},
		{
			name: "Date only layout",
			args: args{
				layout:  "2 January 2006",
				dateStr: "17 December 2021",
			},
			want: time.Date(2021, 12, 17, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "Time only layout",
			args: args{
				layout:  "15:04",
				dateStr: "11:41",
			},
			want: time.Date(now.Year(), now.Month(), now.Day(), 11, 41, 0, 0, time.UTC),
		},
		{
			name: "No year layout",
			args: args{
				layout:  "2 January 15:04",
				dateStr: "17 December 11:41",
			},
			want: time.Date(now.Year(), 12, 17, 11, 41, 0, 0, time.UTC),
		},
		{
			name: "Human parse",
			args: args{
				dateStr: "1 day ago",
			},
			want: time.Now().Add(-24 * time.Hour).Round(time.Second),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := site2rss.TimeParse(tt.args.layout, tt.args.dateStr); !reflect.DeepEqual(got.Round(time.Second), tt.want) {
				t.Errorf("TimeParse() = %v, want %v", got, tt.want)
			}
		})
	}
}
