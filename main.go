package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"rss-generator/providers"
	cacheService "rss-generator/services"

	"github.com/chromedp/chromedp"
)

func main() {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	cache := cacheService.NewMemoryCache()

	http.HandleFunc("/theverge/rss.xml", func(w http.ResponseWriter, r *http.Request) {
		scraper := providers.NewTheVergeScraper(cache)
		xmlStr, err := scraper.Scrape(ctx)
		if err != nil {
			log.Fatal(err)
		}

		w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
		w.Write([]byte(xmlStr))
	})

	fmt.Println("Serving RSS feed at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
