package client

import (
	"log"
	"net/http"
	"os"
	"regexp"
)

func ReadCaddyfile() ([]string, error) {
	fileContent, err := os.ReadFile("/etc/Caddyfile") // Read the text from my caddyFile
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	text := string(fileContent)
	// Define the regular expression pattern
	pattern := regexp.MustCompile(`(?m)\b([a-zA-Z0-9.-]+)\s*{`)

	// Find all matches in the text
	matches := pattern.FindAllStringSubmatch(text, -1)

	// Extract the matched domain names
	var domains []string
	for _, match := range matches {
		if len(match) > 1 {
			domains = append(domains, match[1])
		}
	}
	return domains, nil
}

func QueryURL(url string) bool {
	_, err := http.Get(url)
	if err != nil {
		return false
	}
	return true
}
