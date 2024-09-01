package generator

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
)

type Team struct {
	Name     string
	ImageURL string
}

func GetTeams() ([]Team, error) {
	files, err := os.ReadDir("public/screenshots")
	if err != nil {
		return nil, err
	}

	var teams []Team

	// Add league standings as the first item
	teams = append(teams, Team{
		Name:     "League Standings",
		ImageURL: "/screenshots/league_standings.png",
	})

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".png" && file.Name() != "league_standings.png" {
			name, err := decodeFilename(file.Name())
			if err != nil {
				continue
			}
			teams = append(teams, Team{
				Name:     name,
				ImageURL: "/screenshots/" + file.Name(),
			})
		}
	}

	return teams, nil
}

func decodeFilename(filename string) (string, error) {
	parts := strings.Split(filename, "_")
	encodedPart := strings.TrimSuffix(parts[len(parts)-1], ".png")

	decodedBytes, err := base64.URLEncoding.DecodeString(encodedPart)
	if err != nil {
		return "", err
	}

	return string(decodedBytes), nil
}
