package providers

import (
	"context"
	"encoding/xml"
	"fmt"
	"log"
	"net/url"
	cacheService "rss-generator/services/cache"
	"time"

	"github.com/chromedp/chromedp"
)

const (
	cacheKeyTheVerge = "rss-theverge"
)

// VergeArticle struct to hold scraped data from The Verge
type VergeArticle struct {
	Title   string `json:"title"`
	Link    string `json:"link"`
	Summary string `json:"summary"`
	Date    string `json:"date"`
}

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	PubDate     string    `xml:"pubDate"`
	Items       []RSSItem `xml:"item"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
	Author      string `xml:"author"`
}

type TheVergeScraper struct {
	Cache cacheService.Cacher // Interface for the cache
}

func NewTheVergeScraper(cache cacheService.Cacher) *TheVergeScraper {
	return &TheVergeScraper{Cache: cache}
}

// Scrape scrapes articles from The Verge
func (s *TheVergeScraper) Scrape(ctx context.Context, isJob ...string) (string, error) {
	fmt.Println("Star scraping The Verge...")
	cacheContent, haveCached := s.Cache.Get(cacheKeyTheVerge)
	if haveCached && len(isJob) == 0 {
		fmt.Printf("Hit `%s` cache", cacheKeyTheVerge)
		return cacheContent, nil
	}
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
		return "", err
	}

	xmlStr := generatedTheVergeFeed("The Verge", "https://www.theverge.com/", "Latest articles from The Verge", articles)
	defer func() {
		s.Cache.Set(cacheKeyTheVerge, xmlStr)
	}()

	return xmlStr, nil
}

// parseVergeDate parses dates from The Verge's format.
func parseVergeDate(dateString string) (string, error) {
	var _dateString = dateString
	if len(dateString) <= 19 || dateString[10] != 'T' || dateString[19] != '+' {
		_dateString = dateString + "+00:00"
	}

	s, err := time.Parse(time.RFC3339, _dateString)
	if err == nil {
		str := s.Format(time.DateTime)
		return str, nil
	}

	log.Printf("Error parsing date '%s': %v, using current time", dateString, err)
	return time.Now().Format(time.DateTime), nil
}

func generatedTheVergeFeed(title, link, description string, articles []VergeArticle) string {
	now := time.Now().Format(time.DateTime)
	rss := RSS{
		XMLName: xml.Name{Local: "rss"},
		Version: "2.0",
		Channel: Channel{
			Title:       title,
			Link:        link,
			Description: description,
			PubDate:     now,
			Items:       []RSSItem{},
		},
	}

	for _, article := range articles {
		pubDate := now
		if article.Date != "" {
			pubDate, _ = parseVergeDate(article.Date)
		}

		guidURL, _ := url.Parse(article.Link)
		guid := guidURL.String()

		rssItem := RSSItem{
			Title:       article.Title,
			Link:        article.Link,
			Description: article.Summary,
			PubDate:     pubDate,
			GUID:        guid,
		}
		rss.Channel.Items = append(rss.Channel.Items, rssItem)
	}

	output, err := xml.MarshalIndent(rss, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	result := xml.Header + string(output)
	log.Println("The Verge RSS feed generated successfully.")

	return result
}
