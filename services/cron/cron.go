package cronService

import (
	"context"
	"fmt"
	"log"
	"rss-generator/providers"
	"time"

	"github.com/robfig/cron/v3"
)

type CronService struct {
	cron    *cron.Cron
	scraper providers.Scraper
}

func NewCronService(scraper providers.Scraper) *CronService {
	return &CronService{
		cron:    cron.New(cron.WithSeconds()),
		scraper: scraper,
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

func (s *CronService) SetScraper(scraper providers.Scraper) {
	s.scraper = scraper
}

func (s *CronService) addJob(name string, jobFunc func(ctx context.Context) error) error {
	_, err := s.cron.AddFunc("0 0 0 * * *", func() { // Run every day at 00:00
		log.Printf("Running %s job...", name)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		err := jobFunc(ctx)
		if err != nil {
			log.Printf("Error running %s job: %v", name, err)
		} else {
			log.Printf("%s job completed at %s", name, time.Now().Format(time.DateTime))

		}
	})
	if err != nil {
		return fmt.Errorf("error adding %s job: %w", name, err)
	}
	log.Printf("%s job added to cron.", name)
	return nil

}

func (s *CronService) AddTheVergeJob() error {
	return s.addJob("The Verge", func(ctx context.Context) error {
		_, err := s.scraper.Scrape(ctx, "true")
		return err
	})
}

func (s *CronService) AddFreeCodeCampJob() error {
	return s.addJob("FreeCodeCamp", func(ctx context.Context) error {
		_, err := s.scraper.Scrape(ctx, "true")
		return err
	})
}

func (s *CronService) AddAWSJob() error {
	return s.addJob("AWS-Blog", func(ctx context.Context) error {
		_, err := s.scraper.Scrape(ctx, "true")
		return err
	})
}

func (s *CronService) AddCSSTricksJob() error {
	return s.addJob("CSS-Tricks", func(ctx context.Context) error {
		_, err := s.scraper.Scrape(ctx, "true")
		return err
	})
}
