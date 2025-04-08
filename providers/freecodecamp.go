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
	cacheKeyFreeCodeCamp = "rss-freecodecamp"
)

// FreeCodeCampArticle struct to hold scraped data from FreeCodeCamp
type FreeCodeCampArticle struct {
	Title string `json:"title"`
	Link  string `json:"link"`
	Tag   string `json:"tag"`
	Date  string `json:"date"`
}

type FreeCodeCampScraper struct {
	Cache cacheService.Cacher // Interface for the cache
}

func NewFreeCodeCampScraper(cache cacheService.Cacher) *FreeCodeCampScraper {
	return &FreeCodeCampScraper{Cache: cache}
}

// Scrape scrapes articles from FreeCodeCamp
func (s *FreeCodeCampScraper) Scrape(ctx context.Context, isJob ...string) (string, error) {
	fmt.Println("Star scraping FreeCodeCamp...")
	cacheContent, haveCached := s.Cache.Get(cacheKeyFreeCodeCamp)
	if haveCached && len(isJob) == 0 {
		fmt.Printf("Hit `%s` cache\n", cacheKeyFreeCodeCamp)
		return cacheContent, nil
	}
	var articles []FreeCodeCampArticle

	err := chromedp.Run(ctx,
		chromedp.Navigate("https://www.freecodecamp.org/news/"),
		chromedp.WaitReady(".post-feed"),
		chromedp.Evaluate(`
			let articles = [];
			document.querySelectorAll('.post-feed .post-card').forEach(articleElement => {
				let titleElement = articleElement.querySelector('h2 a');
				let tagElement = articleElement.querySelector('.post-card-tags a');
				let dateElement = articleElement.querySelector('time');
				
				if (titleElement) {
					articles.push({
						title: titleElement.innerText.trim(),
						link: titleElement.href,
						tag: tagElement ? tagElement.innerText.trim() : '',
						date: dateElement ? dateElement.getAttribute('datetime') : '',
					});
				}
			});
			articles;
		`, &articles),
	)

	if err != nil {
		log.Println("Error scraping FreeCodeCamp:", err)
		return "", err
	}

	xmlStr := generatedFreeCodeCampFeed("freeCodeCamp", "https://www.freecodecamp.org/news/", "Latest articles from freeCodeCamp", articles)
	defer func() {
		s.Cache.Set(cacheKeyFreeCodeCamp, xmlStr)
	}()

	return xmlStr, nil
}

// parseFreeCodeCampDate parses dates with timezone
func parseFreeCodeCampDate(dateString string) (string, error) {
	s, err := time.Parse(time.RFC3339, dateString)
	if err != nil {
		fmt.Println("Failed to parse time:", err)
		return dateString, err
	}

	loc, err := time.LoadLocation("Asia/Tokyo") // use (GMT+9) here
	if err != nil {
		fmt.Println("Failed to load timezone:", err)
		return dateString, err
	}
	localTime := s.In(loc)
	// Get the timezone offset in hours
	_, offset := localTime.Zone()
	// Convert seconds to hours
	offsetHours := offset / 3600
	// Format the time as "2006-01-02 15:04:05 GMT+9"
	formattedTime := localTime.Format("2006-01-02 15:04:05") + fmt.Sprintf(" GMT%+d", offsetHours)

	return formattedTime, nil
}

func generatedFreeCodeCampFeed(title, link, description string, articles []FreeCodeCampArticle) string {
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
			pubDate, _ = parseFreeCodeCampDate(article.Date)
		}

		guidURL, _ := url.Parse(article.Link)
		guid := guidURL.String()

		rssItem := RSSItem{
			Title:       article.Title,
			Link:        article.Link,
			Description: article.Tag,
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
	log.Println("FreeCodeCamp RSS feed generated successfully.")

	return result
}
