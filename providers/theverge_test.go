package providers

import (
	"context"
	"encoding/xml"
	"reflect"
	"testing"
	"time"
)

func TestScrapeVerge(t *testing.T) {
	// Create a new context for testing
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run the scraping function
	articles, err := ScrapeVerge(ctx)
	if err != nil {
		t.Fatalf("ScrapeVerge returned an error: %v", err)
	}

	// Check if any articles were returned
	if len(articles) == 0 {
		t.Error("ScrapeVerge returned no articles")
	}

	// Check if the first article has the expected fields
	if len(articles) > 0 {
		firstArticle := articles[0]
		if firstArticle.Title == "" {
			t.Error("First article has an empty title")
		}
		if firstArticle.Link == "" {
			t.Error("First article has an empty link")
		}
		// We can't guarantee summary and date will always be present
	}
}

func TestParseVergeDate(t *testing.T) {
	tests := []struct {
		name        string
		dateString  string
		want        string
		wantErr     bool
		currentTime bool
	}{
		{
			name:       "Valid date with timezone",
			dateString: "2023-10-27T10:00:00+00:00",
			want:       "2023-10-27 10:00:00",
			wantErr:    false,
		},
		{
			name:       "Valid date without timezone",
			dateString: "2023-10-27T10:00:00",
			want:       "2023-10-27 10:00:00",
			wantErr:    false,
		},
		{
			name:        "Invalid date",
			dateString:  "invalid-date",
			wantErr:     false,
			currentTime: true,
		},
		{
			name:       "Empty date",
			dateString: "",
			want:       "",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseVergeDate(tt.dateString)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseVergeDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.currentTime {
				_, err := time.Parse(time.DateTime, got)
				if err != nil {
					t.Errorf("parseVergeDate() returned invalid current time format: %v", err)
				}
			} else if got != tt.want {
				t.Errorf("parseVergeDate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGeneratedTheVergeFeed(t *testing.T) {
	articles := []VergeArticle{
		{
			Title:   "Test Article 1",
			Link:    "https://www.example.com/article1",
			Summary: "Summary of article 1",
			Date:    "2023-10-27T10:00:00+00:00",
		},
		{
			Title:   "Test Article 2",
			Link:    "https://www.example.com/article2",
			Summary: "Summary of article 2",
			Date:    "2023-10-28T12:00:00+00:00",
		},
	}

	feed := GeneratedTheVergeFeed("Test Feed", "https://www.example.com", "Test Description", articles)

	var rss RSS
	err := xml.Unmarshal([]byte(feed), &rss)
	if err != nil {
		t.Fatalf("Failed to unmarshal generated XML: %v", err)
	}

	if rss.Version != "2.0" {
		t.Errorf("Expected RSS version 2.0, got %s", rss.Version)
	}

	if rss.Channel.Title != "Test Feed" {
		t.Errorf("Expected channel title 'Test Feed', got %s", rss.Channel.Title)
	}

	if rss.Channel.Link != "https://www.example.com" {
		t.Errorf("Expected channel link 'https://www.example.com', got %s", rss.Channel.Link)
	}

	if rss.Channel.Description != "Test Description" {
		t.Errorf("Expected channel description 'Test Description', got %s", rss.Channel.Description)
	}

	if len(rss.Channel.Items) != len(articles) {
		t.Errorf("Expected %d items, got %d", len(articles), len(rss.Channel.Items))
	}

	for i, item := range rss.Channel.Items {
		if item.Title != articles[i].Title {
			t.Errorf("Expected item title '%s', got '%s'", articles[i].Title, item.Title)
		}
		if item.Link != articles[i].Link {
			t.Errorf("Expected item link '%s', got '%s'", articles[i].Link, item.Link)
		}
		if item.Description != articles[i].Summary {
			t.Errorf("Expected item description '%s', got '%s'", articles[i].Summary, item.Description)
		}
		expectedDate, _ := parseVergeDate(articles[i].Date)
		if item.PubDate != expectedDate {
			t.Errorf("Expected item pubDate '%s', got '%s'", expectedDate, item.PubDate)
		}
		if item.GUID != articles[i].Link {
			t.Errorf("Expected item GUID '%s', got '%s'", articles[i].Link, item.GUID)
		}
	}
}

func TestRSS_XMLName(t *testing.T) {
	rss := RSS{}
	expected := xml.Name{Local: "rss"}
	if !reflect.DeepEqual(rss.XMLName, expected) {
		t.Errorf("Expected XMLName to be %v, got %v", expected, rss.XMLName)
	}
}
