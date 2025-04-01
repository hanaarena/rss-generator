package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"rss-reader/providers"
	"time"

	"github.com/chromedp/chromedp"
)

// RSS Feed Structure
type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	PubDate     string    `xml:"pubDate"` // RFC822 format
	Items       []RSSItem `xml:"item"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"` // RFC822 format
	GUID        string `xml:"guid"`    // Optional, but good for uniqueness
}

func main() {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var rssFeed RSS

	// Scrape The Verge
	vergeArticles, err := providers.ScrapeVerge(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Create RSS Feed
	rssFeed = createRSSFeed("The Verge", "https://www.theverge.com/", "Latest articles from The Verge", vergeArticles)

	// Marshal to XML
	output, err := xml.MarshalIndent(rssFeed, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(xml.Header + string(output))
	log.Println("RSS feed generated successfully.")

	http.HandleFunc("/rss.xml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.Write([]byte(xml.Header + string(output)))
	})

	fmt.Println("Serving RSS feed at http://localhost:8080/rss.xml")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func createRSSFeed(title, link, description string, articles []providers.VergeArticle) RSS {
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
			pubDate, _ = providers.ParseVergeDate(article.Date)
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

	return rss
}
