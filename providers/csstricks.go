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
	cacheKeyCSSTricks = "rss-css-tricks"
)

type CSSTricksArticle struct {
	Title       string `json:"title"`
	Link        string `json:"link"`
	Description string `json:"description"`
	Date        string `json:"date"`
	Author      string `json:"author"`
	Category    string `json:"category"`
}

type CSSTricksScraper struct {
	Cache cacheService.Cacher
}

func NewCSSTricksScraper(cache cacheService.Cacher) *CSSTricksScraper {
	return &CSSTricksScraper{Cache: cache}
}

// Scrape scrapes articles from CSS-Tricks
func (s *CSSTricksScraper) Scrape(ctx context.Context, isJob ...string) (string, error) {
	fmt.Println("Star scraping CSS-Tricks...")
	cacheContent, haveCached := s.Cache.Get(cacheKeyCSSTricks)
	if haveCached && len(isJob) == 0 {
		fmt.Printf("Hit `%s` cache\n", cacheKeyCSSTricks)
		return cacheContent, nil
	}
	var articles []CSSTricksArticle

	err := chromedp.Run(ctx,
		chromedp.Navigate("https://css-tricks.com/"),
		chromedp.WaitReady(".latest-articles"),
		chromedp.Evaluate(`
			let articles = [];
			document.querySelectorAll('.latest-articles .article-card').forEach(articleElement => {
				let titleElement = articleElement.querySelector('h2 a');
				let descElement = articleElement.querySelector('.article-content');
				let dateElement = articleElement.querySelector('time');
				let authorElement = articleElement.querySelector('.author-row .author-name');
				let tagElement = articleElement.querySelector('.article-article .tags');
				let tagTextArr = [];
				let tags = tagElement.querySelectorAll('a[rel="tag"]');
				tags.forEach(function(tag) {
					tagTextArr.push(tag.textContent);
				});

				if (titleElement) {
					articles.push({
						title: titleElement.innerText.trim(),
						link: titleElement.href,
						description: descElement ? descElement.innerText.trim() : '',
						author: authorElement ? authorElement.innerText.trim() : '',
						date: dateElement ? dateElement.innerText.trim() : '',
						category: tagTextArr.join(', '),
					});
				}
			});
			articles;
		`, &articles),
	)

	if err != nil {
		log.Println("Error scraping CSS-Tricks:", err)
		return "", err
	}

	xmlStr := generatedCSSTricksFeed("CSS-Tricks", "https://css-tricks.com/", "Latest articles from CSS-Tricks", articles)
	defer func() {
		s.Cache.Set(cacheKeyCSSTricks, xmlStr)
	}()

	return xmlStr, nil
}

func generatedCSSTricksFeed(title, link, description string, articles []CSSTricksArticle) string {
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
		pubDate := article.Date
		if pubDate == "" {
			pubDate = now
		}

		guidURL, _ := url.Parse(article.Link)
		guid := guidURL.String()

		rssItem := RSSItem{
			Title:       article.Title,
			Link:        article.Link,
			Description: article.Description,
			PubDate:     pubDate,
			GUID:        guid,
			Author:      article.Author,
			Category:    article.Category,
		}
		rss.Channel.Items = append(rss.Channel.Items, rssItem)
	}

	output, err := xml.MarshalIndent(rss, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	result := xml.Header + string(output)
	log.Println("CSS-Tricks RSS feed generated successfully.")

	return result
}
