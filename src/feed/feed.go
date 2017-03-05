package feed

import (
	"time"
	"github.com/mmcdole/gofeed"
)

type feedItem struct {
	Url         string
	Image       *gofeed.Image
	Title       string
	Description string
	Content     string
	Author      *gofeed.Person
	Published   *time.Time
	Updated     *time.Time
}

// TODO : Fetch images directly from feeds.
func FetchFeedItem(url string) ([]feedItem, error) {
	feeds := make([]feedItem, 1)

	parser := gofeed.NewParser()
	f, err := parser.ParseURL(url)
	if err != nil {
		return feeds, err
	}

	for _, i := range f.Items {
		if i.Link != "" {
			fi := feedItem{
				i.Link, i.Image, i.Title, i.Description, i.Content,
				i.Author, i.PublishedParsed, i.UpdatedParsed,
			}
			feeds = append(feeds, fi)
		}
	}

	return feeds, nil
}
