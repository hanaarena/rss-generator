package providers

import (
	"context"
	"log"
	"time"

	"github.com/chromedp/chromedp"
)

// VergeArticle struct to hold scraped data from The Verge
type VergeArticle struct {
	Title   string `json:"title"`
	Link    string `json:"link"`
	Summary string `json:"summary"`
	Date    string `json:"date"`
}

// ScrapeVerge scrapes articles from The Verge
func ScrapeVerge(ctx context.Context) ([]VergeArticle, error) {
	var articles []VergeArticle

	err := chromedp.Run(ctx,
		chromedp.Navigate("https://www.theverge.com/"),
		chromedp.WaitReady("#content"),
		chromedp.Evaluate(`
			let articles = [];
			document.querySelectorAll('.duet--content-cards--content-card').forEach(articleElement => {
				let titleElement = articleElement.querySelector('a');
				let summaryElement = articleElement.querySelector('.p-dek');
				let dateElement = articleElement.querySelector('.duet--article--timestamp time');

				if (titleElement) {
					articles.push({
						title: titleElement.innerText.trim(),
						link: titleElement.href,
						summary: summaryElement ? summaryElement.innerText.trim() : '',
						date: dateElement ? dateElement.getAttribute('datetime') : '',
					});
				}
			});
			articles;
		`, &articles),
	)

	if err != nil {
		log.Println("Error scraping The Verge:", err)
		return nil, err
	}

	return articles, nil
}

// ParseVergeDate parses dates from The Verge's format.
func ParseVergeDate(dateString string) (string, error) {
	var _dateString string
	if len(dateString) > 19 && dateString[10] == 'T' && dateString[19] == '+' {
		_dateString = dateString
	} else {
		_dateString = dateString+"+00:00"
	}

	s, err := time.Parse(time.RFC3339, _dateString)
	if err == nil {
		str := s.Format(time.DateTime)
		return str, nil
	}

	log.Printf("Error parsing date '%s': %v, using current time", dateString, err)
	return time.Now().Format(time.DateTime), nil
}
