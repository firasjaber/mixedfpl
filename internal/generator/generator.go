package generator

import (
	"encoding/base64"
	"fmt"
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
		fmt.Println("Error reading screenshots directory:", err)
		return nil, err
	}

	var teams []Team

	// Add league standings as the first item
	leagueStandingsFile, err := findLeagueStandingsFile(files)
	if err != nil {
		fmt.Println("Error finding league standings file:", err)
		return nil, err
	}
	teams = append(teams, Team{
		Name:     "League Standings",
		ImageURL: "/screenshots/" + leagueStandingsFile,
	})

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".png" && !strings.HasPrefix(file.Name(), "league_standings") {
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

func findLeagueStandingsFile(files []os.DirEntry) (string, error) {
	for _, file := range files {
		if strings.HasPrefix(file.Name(), "league_standings") && filepath.Ext(file.Name()) == ".png" {
			return file.Name(), nil
		}
	}
	return "", fmt.Errorf("league standings file not found")
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
