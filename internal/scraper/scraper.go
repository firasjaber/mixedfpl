package scraper

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

type Scraper struct {
	LeagueURL string
	busy      sync.Mutex
}

func NewScraper() *Scraper {
	return &Scraper{LeagueURL: "https://fantasy.premierleague.com/leagues/63322/standings/c"}
}

func (s *Scraper) Scrape() error {
	if !s.busy.TryLock() {
		log.Printf("%s: Scraper is busy, skipping this scrape", time.Now().Format(time.RFC3339))
		return fmt.Errorf("scraper is busy, skipping this scrape")
	}
	defer s.busy.Unlock()

	log.Printf("%s: Starting scrape", time.Now().Format(time.RFC3339))
	u := launcher.New().
		Set("no-sandbox", "").
		Set("disable-gpu", "").
		MustLaunch()

	browser := rod.New().ControlURL(u).MustConnect()
	defer browser.MustClose()

	// Navigate to the league page
	page := browser.MustPage("https://fantasy.premierleague.com/leagues/63322/standings/c")
	fmt.Println("Navigated to league page")

	// Wait for the page to load
	page.MustWaitStable()
	time.Sleep(5 * time.Second)

	// take a debug screenshot
	page.MustScreenshot("public/screenshots/debug.png")

	// Select all rows in the standings table
	rows, err := page.Elements("tbody tr[class^='StandingsRow']")
	if err != nil {
		fmt.Println("Error selecting rows:", err)
		return fmt.Errorf("error selecting rows: %v", err)
	}

	fmt.Println("Number of rows:", len(rows))

	var teamLinks []string

	for _, row := range rows {
		// Find the second td in each row
		secondTd, err := row.Element("td:nth-child(2)")
		if err != nil {
			fmt.Println("Error finding second td:", err)
			continue
		}

		// Find the link within the second td
		link, err := secondTd.Element("a")
		if err != nil {
			fmt.Println("Error finding link:", err)
			continue
		}

		// Get the href attribute
		href, err := link.Attribute("href")
		if err != nil {
			fmt.Println("Error getting href:", err)
			continue
		}

		// Add the href to the teamLinks array
		if href != nil {
			teamLinks = append(teamLinks, *href)
		}
	}

	// Log the content of the array
	fmt.Println("Team links:")
	for i, link := range teamLinks {
		fmt.Printf("%d: %s\n", i+1, link)
	}
	fmt.Println("Number of teams:", len(teamLinks))

	// if cookie modal exists, click the close button
	// cookie modal handled := make(chan bool, 1)
	cookieBtn, err := page.Element("#onetrust-accept-btn-handler")
	if err == nil && cookieBtn != nil {
		cookieBtn.MustClick()
		time.Sleep(1 * time.Second)
	}

	// Capture league standings table
	err = s.captureLeagueStandings(page)
	if err != nil {
		fmt.Println("Error capturing league standings:", err)
	}

	// Navigate to each team link and take a screenshot (limited to first 3)
	for _, link := range teamLinks {
		fullURL := "https://fantasy.premierleague.com" + link
		page.MustNavigate(fullURL)
		fmt.Printf("Navigated to: %s\n", fullURL)

		// Wait for the page to load
		page.MustWaitStable()
		time.Sleep(3 * time.Second)

		// Find the heading element and get its text
		headingElement := page.MustElement("h2[class^='Title']")
		headingText, err := headingElement.Text()
		if err != nil {
			fmt.Printf("Error getting heading text for %s: %v\n", link, err)
			headingText = "Unknown" // Use a default value if we can't get the heading
		}

		// Find the element and ensure it's visible
		element := page.MustElement("div[class^='GraphicPatterns__PatternWrapMain']")

		// Adjust z-index to bring the element to the front
		page.MustEval(`() => {
			const element = document.querySelector("div[class^='GraphicPatterns__PatternWrapMain']");
			element.style.zIndex = "9999";
			element.style.position = "relative";
		}`)

		// Set a larger viewport size
		page.MustSetViewport(410, 1200, 1, false)

		// Scroll the element into view
		err = element.ScrollIntoView()
		if err != nil {
			fmt.Printf("Error scrolling to element for %s: %v\n", link, err)
			continue
		}

		// Wait a bit for any animations to complete
		time.Sleep(2 * time.Second)

		// log the heading text
		fmt.Printf("Heading text: %s\n", headingText)

		// Take a screenshot of the element
		filename := generateFilename(link, headingText)
		screenshot, err := element.Screenshot(proto.PageCaptureScreenshotFormatPng, 100)
		if err != nil {
			fmt.Printf("Error taking screenshot for %s: %v\n", link, err)
			continue
		}

		// Save the screenshot to a file
		err = os.WriteFile(filename, screenshot, 0644)
		if err != nil {
			fmt.Printf("Error saving screenshot for %s: %v\n", link, err)
			continue
		}

		fmt.Printf("Screenshot saved: %s\n", filename)

		// Reset zoom
		page.MustEval(`() => {
			document.body.style.zoom = "1";
		}`)
	}

	fmt.Println("Process completed successfully")
	log.Printf("%s: Scrape completed", time.Now().Format(time.RFC3339))
	return nil
}

func (s *Scraper) captureLeagueStandings(page *rod.Page) error {
	// Wait for the standings table to load
	page.MustWaitStable()
	time.Sleep(2 * time.Second)

	// Find the standings table
	element := page.MustElement("div[class^='Layout__Main'] table")

	// Set a larger viewport size
	page.MustSetViewport(410, 1200, 1, false)

	// Scroll the element into view
	err := element.ScrollIntoView()
	if err != nil {
		return fmt.Errorf("error scrolling to standings table: %v", err)
	}

	// Wait a bit for any animations to complete
	time.Sleep(1 * time.Second)

	// Take a screenshot of the element
	screenshot, err := element.Screenshot(proto.PageCaptureScreenshotFormatPng, 100)
	if err != nil {
		return fmt.Errorf("error taking screenshot of standings table: %v", err)
	}

	// Save the screenshot to a file
	filename := filepath.Join("public/screenshots", "league_standings.png")
	err = os.WriteFile(filename, screenshot, 0644)
	if err != nil {
		return fmt.Errorf("error saving standings screenshot: %v", err)
	}

	fmt.Printf("League standings screenshot saved: %s\n", filename)
	return nil
}

func generateFilename(link, headingText string) string {
	// Extract the last part of the link (e.g., "entry/12345/event/1")
	parts := strings.Split(link, "/")
	lastPart := parts[len(parts)-3] // Get the entry number

	// Strip out "Points -" from the heading text
	headingText = strings.TrimPrefix(headingText, "Points - ")

	// Sanitize the heading text for use in a filename
	sanitizedHeading := sanitizeFilename(headingText)

	// Encode the sanitized heading using base64
	encodedHeading := base64.URLEncoding.EncodeToString([]byte(sanitizedHeading))

	// Create a filename with the entry number and encoded heading text saving it in the public/screenshots directory
	return filepath.Join("public/screenshots", fmt.Sprintf("team_%s_%s.png", lastPart, encodedHeading))
}

func sanitizeFilename(name string) string {
	// Replace spaces and some problematic characters with underscores
	replacer := strings.NewReplacer(" ", "_", "/", "_", "\\", "_", ":", "_", "*", "_", "?", "_", "\"", "_", "<", "_", ">", "_", "|", "_")
	name = replacer.Replace(name)

	// Remove any other characters that are not Unicode letters, numbers, or underscores
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_' {
			return r
		}
		return -1
	}, name)
}
