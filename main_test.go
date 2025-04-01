package main

import (
	"context"
	"encoding/xml"
	"net/url"
	"reflect"
	"rss-reader/providers"
	"testing"
	"time"
)

func TestCreateRSSFeed(t *testing.T) {
	testCases := []struct {
		name        string
		title       string
		link        string
		description string
		articles    []providers.VergeArticle
		expected    RSS
	}{
		{
			name:        "Empty Articles",
			title:       "Test Title",
			link:        "https://test.com",
			description: "Test Description",
			articles:    []providers.VergeArticle{},
			expected: RSS{
				XMLName: xml.Name{Local: "rss"},
				Version: "2.0",
				Channel: Channel{
					Title:       "Test Title",
					Link:        "https://test.com",
					Description: "Test Description",
					Items:       []RSSItem{},
				},
			},
		},
		{
			name:        "Single Article",
			title:       "Test Title",
			link:        "https://test.com",
			description: "Test Description",
			articles: []providers.VergeArticle{
				{
					Title:   "Article 1",
					Link:    "https://article1.com",
					Summary: "Summary 1",
					Date:    "2023-10-27T10:00:00Z",
				},
			},
			expected: RSS{
				XMLName: xml.Name{Local: "rss"},
				Version: "2.0",
				Channel: Channel{
					Title:       "Test Title",
					Link:        "https://test.com",
					Description: "Test Description",
					Items: []RSSItem{
						{
							Title:       "Article 1",
							Link:        "https://article1.com",
							Description: "Summary 1",
							GUID:        "https://article1.com",
						},
					},
				},
			},
		},
		{
			name:        "Multiple Articles",
			title:       "Test Title",
			link:        "https://test.com",
			description: "Test Description",
			articles: []providers.VergeArticle{
				{
					Title:   "Article 1",
					Link:    "https://article1.com",
					Summary: "Summary 1",
					Date:    "2023-10-27T10:00:00Z",
				},
				{
					Title:   "Article 2",
					Link:    "https://article2.com",
					Summary: "Summary 2",
					Date:    "2023-10-28T12:00:00Z",
				},
			},
			expected: RSS{
				XMLName: xml.Name{Local: "rss"},
				Version: "2.0",
				Channel: Channel{
					Title:       "Test Title",
					Link:        "https://test.com",
					Description: "Test Description",
					Items: []RSSItem{
						{
							Title:       "Article 1",
							Link:        "https://article1.com",
							Description: "Summary 1",
							GUID:        "https://article1.com",
						},
						{
							Title:       "Article 2",
							Link:        "https://article2.com",
							Description: "Summary 2",
							GUID:        "https://article2.com",
						},
					},
				},
			},
		},
		{
			name:        "Invalid Date",
			title:       "Test Title",
			link:        "https://test.com",
			description: "Test Description",
			articles: []providers.VergeArticle{
				{
					Title:   "Article 1",
					Link:    "https://article1.com",
					Summary: "Summary 1",
					Date:    "Invalid Date",
				},
			},
			expected: RSS{
				XMLName: xml.Name{Local: "rss"},
				Version: "2.0",
				Channel: Channel{
					Title:       "Test Title",
					Link:        "https://test.com",
					Description: "Test Description",
					Items: []RSSItem{
						{
							Title:       "Article 1",
							Link:        "https://article1.com",
							Description: "Summary 1",
							GUID:        "https://article1.com",
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := createRSSFeed(tc.title, tc.link, tc.description, tc.articles)

			// Check basic fields
			if result.Version != tc.expected.Version {
				t.Errorf("Expected version %s, got %s", tc.expected.Version, result.Version)
			}
			if result.Channel.Title != tc.expected.Channel.Title {
				t.Errorf("Expected channel title %s, got %s", tc.expected.Channel.Title, result.Channel.Title)
			}
			if result.Channel.Link != tc.expected.Channel.Link {
				t.Errorf("Expected channel link %s, got %s", tc.expected.Channel.Link, result.Channel.Link)
			}
			if result.Channel.Description != tc.expected.Channel.Description {
				t.Errorf("Expected channel description %s, got %s", tc.expected.Channel.Description, result.Channel.Description)
			}

			// Check number of items
			if len(result.Channel.Items) != len(tc.expected.Channel.Items) {
				t.Fatalf("Expected %d items, got %d", len(tc.expected.Channel.Items), len(result.Channel.Items))
			}

			// Check each item
			for i, item := range result.Channel.Items {
				expectedItem := tc.expected.Channel.Items[i]
				if item.Title != expectedItem.Title {
					t.Errorf("Expected item title %s, got %s", expectedItem.Title, item.Title)
				}
				if item.Link != expectedItem.Link {
					t.Errorf("Expected item link %s, got %s", expectedItem.Link, item.Link)
				}
				if item.Description != expectedItem.Description {
					t.Errorf("Expected item description %s, got %s", expectedItem.Description, item.Description)
				}
				if item.GUID != expectedItem.GUID {
					t.Errorf("Expected item GUID %s, got %s", expectedItem.GUID, item.GUID)
				}
				// We can't check PubDate directly because it's dynamic, but we can check if it's not empty
				if item.PubDate == "" {
					t.Errorf("Expected item PubDate to not be empty")
				}
			}
		})
	}
}

func TestParseVergeDate(t *testing.T) {
	testCases := []struct {
		name          string
		dateString    string
		expectedError bool
	}{
		{
			name:          "Valid ISO 8601",
			dateString:    "2023-10-27T10:00:00Z",
			expectedError: false,
		},
		{
			name:          "Invalid Date",
			dateString:    "Invalid Date",
			expectedError: true,
		},
		{
			name:          "Empty String",
			dateString:    "",
			expectedError: true,
		},
		{
			name:          "Partial ISO 8601",
			dateString:    "2023-10-27",
			expectedError: false,
		},
		{
			name: "Valid ISO 8601 with timezone",
			dateString: "2023-11-20T15:30:00+09:00",
			expectedError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := providers.ParseVergeDate(tc.dateString)
			if tc.expectedError && err == nil {
				t.Errorf("Expected error, but got nil")
			}
			if !tc.expectedError && err != nil {
				t.Errorf("Expected no error, but got %v", err)
			}
		})
	}
}

func TestRSSItem_GUID(t *testing.T) {
	testCases := []struct {
		name     string
		link     string
		expected string
	}{
		{
			name:     "Valid URL",
			link:     "https://example.com/article/123",
			expected: "https://example.com/article/123",
		},
		{
			name:     "Empty URL",
			link:     "",
			expected: "",
		},
		{
			name:     "Invalid URL",
			link:     "://invalid-url",
			expected: "://invalid-url",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			guidURL, _ := url.Parse(tc.link)
			guid := guidURL.String()
			if guid != tc.expected {
				t.Errorf("Expected GUID %s, got %s", tc.expected, guid)
			}
		})
	}
}

func TestRSS_XMLName(t *testing.T) {
	rss := RSS{}
	expected := xml.Name{Local: "rss"}
	if !reflect.DeepEqual(rss.XMLName, expected) {
		t.Errorf("Expected XMLName %v, got %v", expected, rss.XMLName)
	}
}

func TestChannel_XMLFields(t *testing.T) {
	channel := Channel{
		Title:       "Test Title",
		Link:        "https://test.com",
		Description: "Test Description",
		PubDate:     "Test Date",
		Items:       []RSSItem{},
	}

	if channel.Title != "Test Title" {
		t.Errorf("Expected Title to be 'Test Title', got '%s'", channel.Title)
	}
	if channel.Link != "https://test.com" {
		t.Errorf("Expected Link to be 'https://test.com', got '%s'", channel.Link)
	}
	if channel.Description != "Test Description" {
		t.Errorf("Expected Description to be 'Test Description', got '%s'", channel.Description)
	}
	if channel.PubDate != "Test Date" {
		t.Errorf("Expected PubDate to be 'Test Date', got '%s'", channel.PubDate)
	}
	if len(channel.Items) != 0 {
		t.Errorf("Expected Items to be empty, got %d items", len(channel.Items))
	}
}

func TestRSSItem_XMLFields(t *testing.T) {
	item := RSSItem{
		Title:       "Test Title",
		Link:        "https://test.com",
		Description: "Test Description",
		PubDate:     "Test Date",
		GUID:        "Test GUID",
	}

	if item.Title != "Test Title" {
		t.Errorf("Expected Title to be 'Test Title', got '%s'", item.Title)
	}
	if item.Link != "https://test.com" {
		t.Errorf("Expected Link to be 'https://test.com', got '%s'", item.Link)
	}
	if item.Description != "Test Description" {
		t.Errorf("Expected Description to be 'Test Description', got '%s'", item.Description)
	}
	if item.PubDate != "Test Date" {
		t.Errorf("Expected PubDate to be 'Test Date', got '%s'", item.PubDate)
	}
	if item.GUID != "Test GUID" {
		t.Errorf("Expected GUID to be 'Test GUID', got '%s'", item.GUID)
	}
}

func TestScrapeVerge(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	articles, err := providers.ScrapeVerge(ctx)
	if err != nil {
		t.Fatalf("ScrapeVerge failed: %v", err)
	}

	if len(articles) == 0 {
		t.Errorf("Expected at least one article, got 0")
	}

	for _, article := range articles {
		if article.Title == "" {
			t.Errorf("Expected article title to not be empty")
		}
		if article.Link == "" {
			t.Errorf("Expected article link to not be empty")
		}
	}
}
