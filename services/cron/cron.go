package cronService

import (
	"context"
	"fmt"
	"log"
	"rss-generator/providers" // Import your providers package
	"time"

	"github.com/robfig/cron/v3"
)

type CronService struct {
	cron *cron.Cron
	theVergeScraper *providers.TheVergeScraper
}

func NewCronService(theVergeScraper *providers.TheVergeScraper) *CronService {
	return &CronService{
		cron: cron.New(cron.WithSeconds()),
		theVergeScraper: theVergeScraper,
	}
}

func (s *CronService) Start() {
	s.cron.Start()
	log.Println("Cron service started.")
}

func (s *CronService) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	log.Println("Cron service stopped.")
}

func (s *CronService) AddTheVergeJob() error {
	// Run every day at 00:00
	// See https://pkg.go.dev/github.com/robfig/cron/v3#hdr-CRON_Expression_Format
	_, err := s.cron.AddFunc("0 0 0 * * *", func() {
		log.Println("Running The Verge job...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		_, err := s.theVergeScraper.Scrape(ctx, "true")
		if err != nil {
			log.Printf("Error scraping The Verge: %v", err)
		} else {
			log.Printf("The Verge job completed at %s", time.Now().Format(time.DateTime))
		}
	})
	if err != nil {
		return fmt.Errorf("error adding The Verge job: %w", err)
	}
	log.Println("The Verge job added to cron.")
	return nil
}
