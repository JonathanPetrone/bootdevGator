package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"github/jonathanpetrone/bootdevBlogAgg/internal/database"
	"html"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	feed := &RSSFeed{}
	client := &http.Client{}

	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return feed, err
	}

	req.Header.Set("User-Agent", "gator")

	resp, err := client.Do(req)
	if err != nil {
		return feed, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return feed, err
	}

	if err := xml.Unmarshal(body, feed); err != nil {
		return feed, err
	}

	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)

	for i := range feed.Channel.Item {
		feed.Channel.Item[i].Title = html.UnescapeString(feed.Channel.Item[i].Title)
		feed.Channel.Item[i].Description = html.UnescapeString(feed.Channel.Item[i].Description)
	}

	return feed, nil
}

func scrapeFeeds(ctx context.Context, s *state) error {
	// Access the database through `s.db` and get the next feed to fetch
	nextFeed, err := s.db.GetNextFeedToFetch(ctx)
	if err != nil {
		// Check if there are no feeds to fetch
		if err == sql.ErrNoRows {
			log.Printf("No feeds to fetch at the moment.")
			return err
		}
		// Log other types of database errors
		log.Printf("Error fetching next feed: %v", err)
		return nil
	}

	log.Printf("Fetching feed: %s (%s)", nextFeed.Name, nextFeed.Url)

	// Call `fetchFeed` to fetch and parse the feed
	feed, err := fetchFeed(ctx, nextFeed.Url)
	if err != nil {
		log.Printf("Error fetching feed %s: %v", nextFeed.Url, err)
		return err
	}

	// Log or process the feed items
	for _, item := range feed.Channel.Item {
		postID := uuid.New()
		now := time.Now()

		// Parse the pub date (you'll need to handle different date formats)
		pubDate, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			// Try alternative format if the first one fails
			pubDate, err = time.Parse(time.RFC1123, item.PubDate)
			if err != nil {
				log.Printf("Error parsing date %s: %v", item.PubDate, err)
				continue
			}
		}

		// Try to create the post
		_, err = s.db.CreatePost(ctx, database.CreatePostParams{
			ID:        postID,
			CreatedAt: now,
			UpdatedAt: now,
			Title:     item.Title,
			Url:       item.Link,
			Description: sql.NullString{
				String: item.Description,
				Valid:  item.Description != "",
			},
			PublishedAt: pubDate,
			FeedID:      nextFeed.ID,
		})

		if err != nil {
			// Check if it's a duplicate URL error
			if strings.Contains(err.Error(), "duplicate key") {
				continue // skip this post and move on
			}
			log.Printf("Failed to create post: %v", err)
			continue
		}
	}

	// Mark the feed as fetched in the database
	err = s.db.MarkFeedFetched(ctx, database.MarkFeedFetchedParams{
		LastFetchedAt: sql.NullTime{Time: time.Now(), Valid: true},
		UpdatedAt:     time.Now(),
		ID:            nextFeed.ID,
	})
	if err != nil {
		log.Printf("Error updating feed %s: %v", nextFeed.Url, err)
		return err
	}

	log.Printf("Successfully fetched and processed feed: %s", nextFeed.Name)
	return nil
}
