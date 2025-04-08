package providers

import (
	"context"
	"encoding/xml"
	"fmt"
	"log"
	"net/url"
	cacheService "rss-generator/services/cache"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

const (
	cacheKeyAWS = "rss-aws"
)

// AWSArticle struct to hold scraped data from AWS Blogs
type AWSArticle struct {
	Title       string `json:"title"`
	Link        string `json:"link"`
	Description string `json:"description"`
	Author      string `json:"author"`
	Date        string `json:"date"`
}

type AWSSraper struct {
	Cache cacheService.Cacher
}

func NewAWSSraper(cache cacheService.Cacher) *AWSSraper {
	return &AWSSraper{Cache: cache}
}

// Scrape scrapes articles from AWS Blogs
func (s *AWSSraper) Scrape(ctx context.Context, isJob ...string) (string, error) {
	fmt.Println("Star scraping AWS Blogs...")
	cacheContent, haveCached := s.Cache.Get(cacheKeyAWS)
	if haveCached && len(isJob) == 0 {
		fmt.Printf("Hit `%s` cache\n", cacheKeyAWS)
		return cacheContent, nil
	}
	var articles []AWSArticle

	header := network.Headers{
		"Accept-Language": "en-US,en;q=0.9",
	}

	err := chromedp.Run(ctx,
		network.SetExtraHTTPHeaders(header),
		chromedp.Navigate("https://aws.amazon.com/blogs"),
		chromedp.WaitReady(".aws-directories-container-wrapper"),
		chromedp.Evaluate(`
			let articles = [];
			document.querySelectorAll('.aws-directories-container .m-card.m-list-card').forEach(row => {
				let titleElement = row.querySelector('.m-card-title a');
				let descElement = row.querySelector('.m-card-description');
				let infoElement = row.querySelector('.m-card-info');
				let arr = infoElement.innerText.split(',');
				
				if (titleElement) {
					articles.push({
						title: titleElement.innerText.trim(),
						link: titleElement.href,
						description: descElement ? descElement.innerText.trim() : '',
						author: arr[0].trim(),
						date: arr[1].trim()
					});
				}
			});
			articles;
		`, &articles),
	)

	if err != nil {
		log.Println("Error scraping AWS Blogs:", err)
		return "", err
	}

	xmlStr := generatedAWSFeed("AWS Blogs", "https://aws.amazon.com/blogs/", "Latest articles from AWS Blogs", articles)
	defer func() {
		s.Cache.Set(cacheKeyAWS, xmlStr)
	}()

	return xmlStr, nil
}

func generatedAWSFeed(title, link, description string, articles []AWSArticle) string {
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
		guidURL, _ := url.Parse(article.Link)
		guid := guidURL.String()

		rssItem := RSSItem{
			Title:       article.Title,
			Link:        article.Link,
			Description: article.Description,
			Author:      article.Author,
			PubDate:     article.Date,
			GUID:        guid,
		}
		rss.Channel.Items = append(rss.Channel.Items, rssItem)
	}

	output, err := xml.MarshalIndent(rss, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	result := xml.Header + string(output)
	log.Println("AWS Blogs RSS feed generated successfully.")

	return result
}
