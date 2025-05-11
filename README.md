# MixedFPL

A simple web application that I made to see the squads of my friends in a FPL league in one page with frequent updates, and to practise Golang.

You can easily use it yourself by updating the league URL, and simply deploy it with docker-compose.

## Features

- Displays all teams from a private FPL league on a single page
- Uses [go-rod](https://github.com/go-rod/rod) to capture the screenshots with a headless browser
- Automatic updates every 2 minutes
- Live at [fpl.firrj.com](https://fpl.firrj.com)

## How it Works

The application uses a headless browser to periodically scrape team information from the FPL website. The data is then rendered through a web template, making it easy to view all teams at once.

## Running Locally

1. Build and run using Docker:
```bash
docker-compose up --build
```

2. Access the application at `http://localhost:8080`
