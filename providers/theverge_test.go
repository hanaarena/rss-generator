package providers

import (
	"context"
	"errors"
	"fmt"
	cacheService "rss-generator/services/cache"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
)

// MockCache is a mock implementation of the Cacher interface for testing.
type MockCache struct {
	data map[string]string
	err  error
}

func NewMockCache() *MockCache {
	return &MockCache{
		data: make(map[string]string),
	}
}

func (m *MockCache) Get(key string) (string, bool) {
	if m.err != nil {
		return "", false
	}
	value, ok := m.data[key]
	return value, ok
}

func (m *MockCache) Set(key, value string) {
	if m.err != nil {
		return
	}
	m.data[key] = value
}

func (m *MockCache) Delete(key string) {
	delete(m.data, key)
}

func TestNewTheVergeScraper(t *testing.T) {
	cache := cacheService.NewMemoryCache()
	scraper := NewTheVergeScraper(cache)
	assert.NotNil(t, scraper)
	assert.Equal(t, cache, scraper.Cache)
}

func TestTheVergeScraper_Scrape_Cached(t *testing.T) {
	mockCache := NewMockCache()
	mockCache.data[cacheKeyTheVerge] = "cached-xml-data"
	scraper := NewTheVergeScraper(mockCache)

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	result, err := scraper.Scrape(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "cached-xml-data", result)
}

func TestTheVergeScraper_Scrape_NotCached(t *testing.T) {
	mockCache := NewMockCache()
	scraper := NewTheVergeScraper(mockCache)

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	result, err := scraper.Scrape(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "<rss")
	assert.Contains(t, result, "<channel")
	assert.Contains(t, result, "<item")
	assert.Contains(t, result, "The Verge")
	assert.Contains(t, result, "https://www.theverge.com/")
	assert.Contains(t, result, "<title>")
	assert.Contains(t, result, "<link>")
	assert.Contains(t, result, "<description>")
	assert.Contains(t, result, "<pubDate>")
	assert.Contains(t, result, "<guid>")
	assert.Equal(t, result, mockCache.data[cacheKeyTheVerge])
}

func TestTheVergeScraper_Scrape_Error(t *testing.T) {
	mockCache := NewMockCache()
	scraper := NewTheVergeScraper(mockCache)

	// Create a context with a very short timeout to force an error
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	result, err := scraper.Scrape(ctx)
	assert.Error(t, err)
	assert.Empty(t, result)
}

func TestTheVergeScraper_Scrape_SetCacheError(t *testing.T) {
	mockCache := NewMockCache()
	mockCache.err = errors.New("set cache error")
	scraper := NewTheVergeScraper(mockCache)

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	result, err := scraper.Scrape(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.NotContains(t, mockCache.data, cacheKeyTheVerge)
}

func TestParseVergeDate(t *testing.T) {
	testCases := []struct {
		name           string
		dateString     string
		expected       string
		expectingError bool
	}{
		{
			name:           "Valid Date",
			dateString:     "2025-04-02T13:05:50+00:00",
			expected:       "2025-04-02 13:05:50",
			expectingError: false,
		},
		{
			name:           "Invalid Date",
			dateString:     "invalid-date",
			expected:       time.Now().Format(time.DateTime),
			expectingError: false,
		},
		{
			name:           "Date without timezone",
			dateString:     "2023-10-28T10:00:00",
			expected:       "2023-10-28 10:00:00",
			expectingError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parseVergeDate(tc.dateString)
			if tc.expectingError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tc.name == "Invalid Date" {
					// For invalid date, we only check the format, not the exact time
					_, err := time.Parse(time.DateTime, result)
					assert.NoError(t, err)
				} else {
					assert.Equal(t, tc.expected, result)
				}
			}
		})
	}
}

func TestGeneratedTheVergeFeed(t *testing.T) {
	articles := []VergeArticle{
		{
			Title:   "Test Article 1",
			Link:    "https://www.example.com/article1",
			Summary: "Summary 1",
			Date:    "2023-10-27T10:00:00+00:00",
		},
		{
			Title:   "Test Article 2",
			Link:    "https://www.example.com/article2",
			Summary: "Summary 2",
			Date:    "2023-10-28T12:00:00+00:00",
		},
	}

	xmlStr := generatedTheVergeFeed("Test Feed", "https://www.example.com", "Test Description", articles)
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
