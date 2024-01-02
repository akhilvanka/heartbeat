package client

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

type Payload struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	SystemName string             `bson:"system_name,omitempty"`
	UploadTime int64              `bson:"upload_time,omitempty"`
	Services   []Domain           `bson:"services,omitempty"`
}

type Domain struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	ServiceName   string             `bson:"service_name,omitempty"`
	ServiceStatus bool               //`bson:"service_status,omitempty"`
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

func CollectionRun() (Payload, error) {
	var statusList []Domain
	data, err := ReadCaddyfile()
	if err != nil {
		return Payload{}, err
	}
	for _, domain := range data {
		status := QueryURL(fmt.Sprintf("https://%s", domain))
		if status {
			log.Printf("Domain %s is up", domain)
			a := Domain{
				ServiceName:   domain,
				ServiceStatus: true,
			}
			statusList = append(statusList, a)
		} else {
			log.Printf("Domain %s is down", domain)
			a := Domain{
				ServiceName:   domain,
				ServiceStatus: false,
			}
			statusList = append(statusList, a)
		}
	}
	hostName, err := os.Hostname()
	payloadMessage := Payload{
		SystemName: hostName,
		UploadTime: time.Now().Unix(),
		Services:   statusList,
	}
	return payloadMessage, nil
}

func SendData(URI string, data Payload) error {

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(URI).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		return err
	}
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()
	// Send a ping to confirm a successful connection
	if err := client.Database("Mirror").RunCommand(context.TODO(), bson.D{{"ping", 1}}).Err(); err != nil {
		return err
	}

	collection := client.Database("Mirror").Collection("data")
	hostName, _ := os.Hostname()
	filter := bson.D{{"system_name", hostName}}
	_, err = collection.ReplaceOne(context.TODO(), filter, &data, options.Replace().SetUpsert(true))
	log.Printf("Sending data packet to database")
	if err != nil {
		return err
	}
	return nil
}
