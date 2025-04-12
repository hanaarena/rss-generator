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
	cacheKeyNodeWeekly = "rss-nodeweekly"
)

type NodeWeeklyArticle struct {
	Title   string `json:"title"`
	Link    string `json:"link"`
	PubDate string `json:"pubDate"`
}

type NodeWeeklyScraper struct {
	Cache cacheService.Cacher
}

func NewNodeWeeklyScraper(cache cacheService.Cacher) *NodeWeeklyScraper {
	return &NodeWeeklyScraper{Cache: cache}
}

// Scrape scrapes articles from the Node Weekly issue
func (s *NodeWeeklyScraper) Scrape(ctx context.Context, isJob ...string) (string, error) {
	fmt.Println("Star scraping Node Weekly...")
	cacheContent, haveCached := s.Cache.Get(cacheKeyNodeWeekly)
	if haveCached && len(isJob) == 0 {
		fmt.Printf("Hit `%s` cache\n", cacheKeyNodeWeekly)
		return cacheContent, nil
	}
	var articles []NodeWeeklyArticle

	err := chromedp.Run(ctx,
		chromedp.Navigate("https://nodeweekly.com/issues"),
		chromedp.WaitReady(".contained"),
		chromedp.Evaluate(`
			let articles = [];
			document.querySelectorAll('.issues .issue').forEach(item => {
				let title = '';
				let link = '';
				let pubDate = '';

				const tiitleElement = item.querySelector('a');
				title = tiitleElement ? tiitleElement.innerText : '';
				link = tiitleElement ? 'https://nodeweekly.com/' + tiitleElement.href : '';
				const dateElement = tiitleElement.nextSibling;
				pubDate = dateElement ? dateElement.nodeValue.replace(" — ", '') : '';
				articles.push({ title, link, pubDate });
			});
			articles;
		`, &articles),
	)

	if err != nil {
		log.Println("Error scraping Node Weekly:", err)
		return "", err
	}

	xmlStr := generatedNodeWeeklyFeed("Node Weekly", "https://nodeweekly.com/", "A free, once–weekly round-up of Node.js news and articles.", articles)
	defer func() {
		s.Cache.Set(cacheKeyNodeWeekly, xmlStr)
	}()

	return xmlStr, nil
}

func generatedNodeWeeklyFeed(title, link, description string, articles []NodeWeeklyArticle) string {
	rss := RSS{
		XMLName: xml.Name{Local: "rss"},
		Version: "2.0",
		Channel: Channel{
			Title:       title,
			Link:        link,
			Description: description,
			PubDate:     time.Now().Format(time.RFC1123Z),
			Items:       []RSSItem{},
		},
	}

	for _, article := range articles {
		guidURL, _ := url.Parse(article.Link)
		guid := guidURL.String()

		rssItem := RSSItem{
			Title:   article.Title,
			Link:    article.Link,
			PubDate: article.PubDate,
			GUID:    guid,
		}
		rss.Channel.Items = append(rss.Channel.Items, rssItem)
	}

	output, err := xml.MarshalIndent(rss, "", "  ")
	if err != nil {
		log.Printf("Error marshalling Node Weekly RSS feed: %v", err)
		return ""
	}

	result := xml.Header + string(output)
	log.Println("Node Weekly RSS feed generated successfully.")

	return result
}
