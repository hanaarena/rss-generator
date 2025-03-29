package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
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

	var articles []Article
	var rssFeed RSS

	err := chromedp.Run(ctx,
		chromedp.Navigate("https://www.theverge.com/"),
		chromedp.WaitReady("#content"), // Wait for the article list to load (adjust selector if needed)
		chromedp.Evaluate(`
			let articles = [];
			document.querySelectorAll('.duet--content-cards--content-card').forEach(articleElement => { // Adjust selector if needed
				let titleElement = articleElement.querySelector('a'); // Adjust selector if needed
				let summaryElement = articleElement.querySelector('.p-dek'); // Adjust selector if needed
				let dateElement = articleElement.querySelector('.duet--article--timestamp time'); // Adjust selector if needed

				if (titleElement) {
					articles.push({
						title: titleElement.innerText.trim(),
						link: titleElement.href,
						summary: '',
						date: dateElement ? dateElement.getAttribute('datetime') : '',
					});
				}
			});
			articles;
		`, &articles),
	)

	if err != nil {
		log.Fatal(err)
	}

	// Create RSS Feed
	rssFeed = createRSSFeed("The Verge", "https://www.theverge.com/", "Latest articles from The Verge", articles)

	// Marshal to XML
	output, err := xml.MarshalIndent(rssFeed, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(xml.Header + string(output)) // Print RSS XML to console (you can save to a file or serve it)
	log.Println("RSS feed generated successfully.")

	http.HandleFunc("/rss.xml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.Write([]byte(xml.Header + string(output))) // Serve the generated XML
	})

	fmt.Println("Serving RSS feed at http://localhost:8080/rss.xml")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Article struct to hold scraped data
type Article struct {
	Title   string `json:"title"`
	Link    string `json:"link"`
	Summary string `json:"summary"`
	Date    string `json:"date"` // Could be string or time.Time depending on parsing
}

func createRSSFeed(title, link, description string, articles []Article) RSS {
	now := time.Now().Format(time.RFC822) // RFC822 date format for RSS

	rss := RSS{
		XMLName: xml.Name{Local: "rss"},
		Version: "2.0",
		Channel: Channel{
			Title:       title,
			Link:        link,
			Description: description,
			PubDate:     now, // Channel publication date (can be current time)
			Items:       []RSSItem{},
		},
	}

	for _, article := range articles {
		pubDate := now // Default to current time if article date parsing fails
		if article.Date != "" {
			// Try to parse the date from the website (adjust format as needed)
			parsedTime, err := parseVergeDate(article.Date) // Custom date parsing function
			if err == nil {
				pubDate = parsedTime.Format(time.RFC822)
			} else {
				log.Printf("Error parsing date '%s': %v, using current time", article.Date, err)
			}
		}

		// Create GUID (unique identifier for the item, using the article link is often sufficient)
		guidURL, _ := url.Parse(article.Link) // Handle potential parsing errors
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

// parseVergeDate is a placeholder function to parse dates from The Verge's format.
// You'll need to implement this based on how dates are presented on the website.
// Example: "2 hours ago", "Yesterday", "Oct 26, 2023", etc.
// This is a simplified example and might need more robust parsing.
func parseVergeDate(dateString string) (time.Time, error) {
	// Example: If dates are in "YYYY-MM-DDTHH:mm:ssZ" format (ISO 8601)
	if strings.Contains(dateString, "T") { // Heuristic to check for ISO format
		return time.Parse(time.RFC3339, dateString)
	}
	// Add more parsing logic for other date formats you find on The Verge
	return time.Now(), fmt.Errorf("unsupported date format: %s", dateString) // Default to current time and error
}
