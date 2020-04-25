package feed

import (
	ctxt "context"
	"fmt"
	"sync"
	"time"

	"github.com/mmcdole/gofeed"

	"github.com/Necoro/feed2imap-go/internal/log"
)

func context() (ctxt.Context, ctxt.CancelFunc) {
	return ctxt.WithTimeout(ctxt.Background(), 60*time.Second)
}

func parseFeed(feed *Feed) error {
	ctx, cancel := context()
	defer cancel()
	fp := gofeed.NewParser()
	parsedFeed, err := fp.ParseURLWithContext(feed.Url, ctx)
	if err != nil {
		return fmt.Errorf("while fetching %s from %s: %w", feed.Name, feed.Url, err)
	}

	feed.feed = parsedFeed
	feed.items = make([]feeditem, len(parsedFeed.Items))
	for idx, item := range parsedFeed.Items {
		feed.items[idx] = feeditem{parsedFeed, item}
	}
	return nil
}

func handleFeed(feed *Feed, group *sync.WaitGroup) {
	defer group.Done()
	log.Printf("Fetching %s from %s", feed.Name, feed.Url)

	err := parseFeed(feed)
	if err != nil {
		log.Error(err)
	}
}

func (feeds Feeds) Parse() int {
	feeds.ForeachGo(handleFeed)

	ctr := 0
	for _, feed := range feeds.feeds {
		success := feed.Success()
		feed.cached.Checked(!success)

		if success {
			ctr++
		}
	}

	return ctr
}
