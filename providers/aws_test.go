package providers

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
)

func TestNewAWSSraper(t *testing.T) {
	cache := NewMockCache()
	scraper := NewAWSSraper(cache)
	assert.NotNil(t, scraper)
	assert.Equal(t, cache, scraper.Cache)
}

func TestAWSSraper_Scrape_Cached(t *testing.T) {
	mockCache := NewMockCache()
	mockCache.data[cacheKeyAWS] = "cached-xml-data"
	scraper := NewAWSSraper(mockCache)

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	result, err := scraper.Scrape(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "cached-xml-data", result)
}

func TestAWSSraper_Scrape_NotCached(t *testing.T) {
	mockCache := NewMockCache()
	scraper := NewAWSSraper(mockCache)

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	result, err := scraper.Scrape(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "<rss")
	assert.Contains(t, result, "<channel")
	assert.Contains(t, result, "<item")
	assert.Contains(t, result, "AWS Blogs")
	assert.Contains(t, result, "https://aws.amazon.com/blogs/")
	assert.Contains(t, result, "<title>")
	assert.Contains(t, result, "<link>")
	assert.Contains(t, result, "<description>")
	assert.Contains(t, result, "<pubDate>")
	assert.Contains(t, result, "<guid>")
	assert.Equal(t, result, mockCache.data[cacheKeyAWS])
}

func TestAWSSraper_Scrape_Error(t *testing.T) {
	mockCache := NewMockCache()
	scraper := NewAWSSraper(mockCache)

	// Create a context with a very short timeout to force an error
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	result, err := scraper.Scrape(ctx)
	assert.Error(t, err)
	assert.Empty(t, result)
}

func TestAWSSraper_Scrape_SetCacheError(t *testing.T) {
	mockCache := NewMockCache()
	mockCache.err = errors.New("set cache error")
	scraper := NewAWSSraper(mockCache)

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	result, err := scraper.Scrape(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.NotContains(t, mockCache.data, cacheKeyAWS)
}

func TestGeneratedAWSFeed(t *testing.T) {
	articles := []AWSArticle{
		{
			Title:       "Test Article 1",
			Link:        "https://www.example.com/article1",
			Description: "Summary 1",
			Date:        "October 27, 2023",
		},
		{
			Title:       "Test Article 2",
			Link:        "https://www.example.com/article2",
			Description: "Summary 2",
			Date:        "Jan 2, 2023",
		},
	}

	xmlStr := generatedAWSFeed("Test Feed", "https://www.example.com", "Test Description", articles)
	fmt.Println(xmlStr)
	assert.NotEmpty(t, xmlStr)
	assert.Contains(t, xmlStr, "<rss")
	assert.Contains(t, xmlStr, "<channel")
	assert.Contains(t, xmlStr, "<item")
	assert.Contains(t, xmlStr, "Test Feed")
	assert.Contains(t, xmlStr, "https://www.example.com")
	assert.Contains(t, xmlStr, "Test Description")
	assert.Contains(t, xmlStr, "<title>Test Article 1</title>")
	assert.Contains(t, xmlStr, "<link>https://www.example.com/article1</link>")
	assert.Contains(t, xmlStr, "<description>Summary 1</description>")
	assert.Contains(t, xmlStr, "<title>Test Article 2</title>")
	assert.Contains(t, xmlStr, "<link>https://www.example.com/article2</link>")
	assert.Contains(t, xmlStr, "<description>Summary 2</description>")
}
