package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"rss-generator/providers"
	cacheService "rss-generator/services/cache"
	cronService "rss-generator/services/cron"

	"github.com/chromedp/chromedp"
)

func main() {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	cache := cacheService.NewMemoryCache()

	// start the cron service
	theVergeScraper := providers.NewTheVergeScraper(cache)
	cronService := cronService.NewCronService(theVergeScraper)
	err := cronService.AddTheVergeJob()
	if err != nil {
		log.Fatalf("Failed to add The Verge job: %v", err)
	}
	cronService.Start()

	http.HandleFunc("/theverge/rss.xml", func(w http.ResponseWriter, r *http.Request) {
		scraper := providers.NewTheVergeScraper(cache)
		xmlStr, err := scraper.Scrape(ctx)
		if err != nil {
			log.Fatal(err)
		}

		w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
		w.Write([]byte(xmlStr))
	})

	http.HandleFunc("/freecodecamp/rss.xml", func(w http.ResponseWriter, r *http.Request) {
		scraper := providers.NewFreeCodeCampScraper(cache)
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
