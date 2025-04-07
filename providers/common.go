package providers

import "context"

type Scraper interface {
	Scrape(ctx context.Context, isJob ...string) (string, error)
}
