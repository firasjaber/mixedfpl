package web

import (
	"html/template"
	"net/http"
	"os"
	"strings"

	"github.com/firasjaber/mixedfpl/internal/generator"
	"github.com/firasjaber/mixedfpl/internal/scraper"
)

type Server struct {
	scraper   *scraper.Scraper
	templates *template.Template
}

func NewServer(s *scraper.Scraper) *Server {
	return &Server{
		scraper:   s,
		templates: template.Must(template.ParseFiles("public/template.html")),
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		s.handleIndex(w, r)
	case "/refresh":
		s.handleRefresh(w, r)
	default:
		http.FileServer(http.Dir("public")).ServeHTTP(w, r)
	}
}

func (s *Server) handleIndex(w http.ResponseWriter, _ *http.Request) {
	teams, err := generator.GetTeams()
	if err != nil {
		http.Error(w, "Error getting teams", http.StatusInternalServerError)
		return
	}

	// get the last updated time from the league standings screenshot file name
	// the file name is in the format of league_standings_YYYY-MM-DD_HH-MM-SS.png
	// so we can get the last updated time by getting the file name and parsing it
	files, err := os.ReadDir("public/screenshots")
	if err != nil {
		http.Error(w, "Error reading screenshots directory", http.StatusInternalServerError)
		return
	}

	lastUpdated := ""
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fileName := file.Name()
		if strings.HasPrefix(fileName, "league_standings_") {
			// remove the prefix and the .png from the end
			lastUpdated = fileName[len("league_standings_") : len(fileName)-len(".png")]
			break
		}
	}

	data := struct {
		Teams       []generator.Team
		LastUpdated string
	}{
		Teams:       teams,
		LastUpdated: lastUpdated,
	}

	s.templates.ExecuteTemplate(w, "template.html", data)
}

func (s *Server) handleRefresh(w http.ResponseWriter, _ *http.Request) {
	err := s.scraper.Scrape()
	if err != nil {
		http.Error(w, "Error refreshing data", http.StatusInternalServerError)
		return
	}
}
