package scrape

import (
	"time"

	readability "github.com/go-shiori/go-readability"
)

type ScrapeCtx struct{}

func (s *ScrapeCtx) Scrape(url string) (string, error) {
	r := ""
	article, err := readability.FromURL(url, 30*time.Second)
	if err != nil {
		return r, err
	}
	return article.TextContent, nil
}
