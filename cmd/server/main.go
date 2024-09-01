package main

import (
	"log"
	"net/http"
	"time"

	"github.com/firasjaber/mixedfpl/internal/scraper"
	"github.com/firasjaber/mixedfpl/internal/web"
)

func main() {
	s := scraper.NewScraper()
	// the initial scrape
	s.Scrape()
	// scrape the data every 2 mins
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			s.Scrape() // The scraper now handles its own busy state
		}
	}()
	srv := web.NewServer(s)

	log.Printf("%s: Server starting on :8080", time.Now().Format(time.RFC3339))
	log.Fatal(http.ListenAndServe(":8080", srv))
}
