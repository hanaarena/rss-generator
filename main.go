package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"rss-generator/providers"

	"github.com/chromedp/chromedp"
)

func main() {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	http.HandleFunc("/theverge/rss.xml", func(w http.ResponseWriter, r *http.Request) {
		vergeArticles, err := providers.ScrapeVerge(ctx)
		if err != nil {
			log.Fatal(err)
		}
		res := providers.GeneratedTheVergeFeed("The Verge", "https://www.theverge.com/", "Latest articles from The Verge", vergeArticles)

		w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
		w.Write([]byte(res))
	})

	fmt.Println("Serving RSS feed at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
