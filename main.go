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
	vergeScraper := providers.NewTheVergeScraper(cache)
	freeCodeCampScraper := providers.NewFreeCodeCampScraper(cache)
	cronService := cronService.NewCronService(vergeScraper)
	err := cronService.AddTheVergeJob()
	if err != nil {
		log.Fatalf("Failed to add The Verge job: %v", err)
	}
	cronService.SetScraper(freeCodeCampScraper)
	err = cronService.AddFreeCodeCampJob()
	if err != nil {
		log.Fatalf("Failed to add FreeCodeCamp job: %v", err)
	}
	cronService.Start()

	// Run the job when server up
	scrapers := []struct {
		name    string
		scraper providers.Scraper
	}{{name: "The Verge", scraper: vergeScraper}, {name: "FreeCodeCamp", scraper: freeCodeCampScraper}}
	for _, s := range scrapers {
		log.Printf("Running %s job immediately on startup...", s.name)
		if _, err := s.scraper.Scrape(ctx); err != nil {
			log.Printf("Error running %s job on startup: %v", s.name, err)
		} else {
			log.Printf("%s job completed successfully on startup.", s.name)
		}
	}

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
