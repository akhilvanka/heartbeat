package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type Payload struct {
	SystemStatus int
	Services     []Domain
}

type Domain struct {
	ServiceName   string
	ServiceStatus int
}

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
		if len(match) > 1 && strings.IndexByte(match[1], '.') != -1 {
			domains = append(domains, match[1])
		}
	}
	return domains, nil
}

func QueryURL(url string) bool {
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		return false
	}
	return true
}

func CollectionRun(serverURL string) error {
	var statusList []Domain
	data, err := ReadCaddyfile()
	if err != nil {
		return err
	}
	for _, domain := range data {
		status := QueryURL(fmt.Sprintf("https://%s", domain))
		if status {
			log.Printf("Domain %s is up", domain)
			a := Domain{
				ServiceName:   domain,
				ServiceStatus: 1,
			}
			statusList = append(statusList, a)
		} else {
			log.Printf("Domain %s is down", domain)
			a := Domain{
				ServiceName:   domain,
				ServiceStatus: 0,
			}
			statusList = append(statusList, a)
		}
	}
	payloadMessage := Payload{
		SystemStatus: 1,
		Services:     statusList,
	}
	body, _ := json.Marshal(payloadMessage)

	// Send JSON status
	resp, err := http.Post(serverURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Error found when closing body")
		}
	}(resp.Body) // Handle this and check if it actually leaks (Shouldn't but just to be safe)
	if resp.StatusCode != 200 {
		return errors.New("unable to post data to server")
	}
	return nil
}
