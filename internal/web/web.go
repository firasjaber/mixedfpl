package web

import (
	"html/template"
	"net/http"
	"time"

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

	data := struct {
		Teams       []generator.Team
		LastUpdated time.Time
	}{
		Teams:       teams,
		LastUpdated: time.Now(),
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
