package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/gocolly/colly"
)

type Link struct {
	URL  string `json:"url"`
	Name string `json:"name,omitempty"`
}

type Tooltip struct {
	Name  string `json:"name"`
	Links []Link `json:"links"`
}

type Episode struct {
	Title    string    `json:"title"`
	Tooltips []Tooltip `json:"tooltips"`
}

func main() {
	c := colly.NewCollector()

	var allEpisodes []Episode

	for i := 1; i <= 55; i++ {
		c.OnResponse(func(r *colly.Response) {
			content := string(r.Body)
			lines := strings.Split(content, "\n")

			var episode Episode
			currentTooltip := Tooltip{}

			linkRegex := regexp.MustCompile(`\[(.*?)\]\((.*?)\)`)

			for _, line := range lines {
				if strings.HasPrefix(line, "title: ") {
					episode.Title = strings.TrimPrefix(line, "title: ")
				} else if strings.HasPrefix(line, "## Tooltips") {
					// Ignore, start of section
				} else if strings.HasPrefix(line, "### ") {
					// New person
					if currentTooltip.Name != "" {
						episode.Tooltips = append(episode.Tooltips, currentTooltip)
					}
					currentTooltip = Tooltip{
						Name: strings.TrimPrefix(line, "### "),
					}
				} else if strings.HasPrefix(line, "- ") {
					// New link
					linkText := strings.TrimPrefix(line, "- ")

					link := Link{}
					matches := linkRegex.FindStringSubmatch(linkText)
					if len(matches) == 3 {
						link.Name = matches[1]
						link.URL = matches[2]
					} else if strings.Contains(linkText, "http") {
						link.URL = linkText
					} else {
						link.Name = linkText
					}

					currentTooltip.Links = append(currentTooltip.Links, link)
				}
			}

			// Add the last tooltip
			if currentTooltip.Name != "" {
				episode.Tooltips = append(episode.Tooltips, currentTooltip)
			}

			allEpisodes = append(allEpisodes, episode)
		})

		c.Visit(fmt.Sprintf("https://raw.githubusercontent.com/devtools-fm/devtools.fm/main/pages/episode/%d.mdx", i))
	}

	// Convert struct to JSON
	jsonData, err := json.MarshalIndent(allEpisodes, "", "    ")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Write JSON data to a file
	file, err := os.Create("allEpisodes.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	file.Write(jsonData)
	fmt.Println("Wrote file: allEpisodes.json")
}
