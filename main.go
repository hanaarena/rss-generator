package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"rss-generator/providers"
	cacheService "rss-generator/services/cache"
	cronService "rss-generator/services/cron"
	"strings"

	"github.com/chromedp/chromedp"
)

func main() {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	cache := cacheService.NewMemoryCache()

	// start the cron service
	vergeScraper := providers.NewTheVergeScraper(cache)
	freeCodeCampScraper := providers.NewFreeCodeCampScraper(cache)
	awsScraper := providers.NewAWSSraper(cache)
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
	cronService.SetScraper(awsScraper)
	err = cronService.AddAWSJob()
	if err != nil {
		log.Fatalf("Failed to add AWS job: %v", err)
	}
	cronService.Start()

	// Run the job when server up
	scrapers := []struct {
		name    string
		scraper providers.Scraper
	}{
		{name: "The Verge", scraper: vergeScraper},
		{name: "FreeCodeCamp", scraper: freeCodeCampScraper},
		{name: "AWS", scraper: awsScraper},
	}
	for _, s := range scrapers {
		log.Printf("Running %s job immediately on startup...", s.name)
		if _, err := s.scraper.Scrape(ctx); err != nil {
			log.Printf("Error running %s job on startup: %v", s.name, err)
		} else {
			log.Printf("%s job completed successfully on startup.", s.name)
		}
	}

	// Define a map of provider names to scraper factories
	scraperFactories := map[string]func(cacheService.Cacher) providers.Scraper{
		"theverge":     func(cache cacheService.Cacher) providers.Scraper { return providers.NewTheVergeScraper(cache) },
		"freecodecamp": func(cache cacheService.Cacher) providers.Scraper { return providers.NewFreeCodeCampScraper(cache) },
		"aws":          func(cache cacheService.Cacher) providers.Scraper { return providers.NewAWSSraper(cache) },
	}

	http.HandleFunc("/feed/", func(w http.ResponseWriter, r *http.Request) {
		// Extract the provider name from the URL path
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) >= 4 && parts[len(parts)-1] == "rss.xml" {
			providerName := parts[len(parts)-2]

			// Check if the provider is supported
			factory, ok := scraperFactories[providerName]
			if !ok {
				http.NotFound(w, r)
				return
			}

			// Create the scraper for the provider
			scraper := factory(cache)
			xmlStr, err := scraper.Scrape(ctx)
			if err != nil {
				log.Printf("Error scraping %s: %v", providerName, err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			// Write the RSS XML to the response
			w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
			w.Write([]byte(xmlStr))
			return
		}
		http.NotFound(w, r)
	})

	fmt.Println("Running server at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
