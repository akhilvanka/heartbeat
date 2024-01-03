package server

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	client2 "heartbeat/client"
	"io"
	"log"
	"net/http"
	"time"
)

type Data struct {
	Available int `json:"available"`
	Total     int `json:"total"`
}

func quickServices(cfg client2.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		serverAPI := options.ServerAPI(options.ServerAPIVersion1)
		opts := options.Client().ApplyURI(cfg.Database.URI).SetServerAPIOptions(serverAPI)

		client, err := mongo.Connect(context.TODO(), opts)
		if err != nil {
			log.Printf("Error connecting to DB")
		}
		defer func() {
			if err = client.Disconnect(context.TODO()); err != nil {
				panic(err)
			}
		}()
		// Send a ping to confirm a successful connection
		if err := client.Database("Mirror").RunCommand(context.TODO(), bson.D{{"ping", 1}}).Err(); err != nil {
			log.Printf("Database not responding")
		}

		collection := client.Database("Mirror").Collection("data")
		cur, err := collection.Find(context.TODO(), bson.D{{}})
		if err != nil {
			log.Printf("Unable to query documents")
		}
		available := 0
		total := 0
		for cur.Next(context.TODO()) {
			var elem client2.Payload
			err := cur.Decode(&elem)
			if err != nil {
				log.Printf("Unable to decode documents")
			}
			total = total + len(elem.Services)
			for _, service := range elem.Services {
				if service.ServiceStatus {
					available = available + 1
				}
			}
		}
		if err := cur.Err(); err != nil {
			log.Fatal(err)
		}
		err = cur.Close(context.TODO())
		if err != nil {
			return
		}
		payloadMessage := Data{
			Available: available,
			Total:     total,
		}
		w.Header().Set("Content-Type", "application/json")
		jsonData, err := json.Marshal(payloadMessage)
		log.Printf("%s", string(jsonData))
		if err != nil {
			log.Printf("Error here")
		}
		_, err = w.Write(jsonData)
		if err != nil {
			return
		}
	}
}

func quickServers(cfg client2.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		serverAPI := options.ServerAPI(options.ServerAPIVersion1)
		opts := options.Client().ApplyURI(cfg.Database.URI).SetServerAPIOptions(serverAPI)

		client, err := mongo.Connect(context.TODO(), opts)
		if err != nil {
			log.Printf("Error connecting to DB")
		}
		defer func() {
			if err = client.Disconnect(context.TODO()); err != nil {
				panic(err)
			}
		}()
		// Send a ping to confirm a successful connection
		if err := client.Database("Mirror").RunCommand(context.TODO(), bson.D{{"ping", 1}}).Err(); err != nil {
			log.Printf("Database not responding")
		}

		collection := client.Database("Mirror").Collection("data")
		estCount, estCountErr := collection.EstimatedDocumentCount(context.TODO())
		if estCountErr != nil {
			panic(estCountErr)
		}

		filter := bson.D{{"upload_time", bson.D{{"$gt", time.Now().Unix() - 600}}}}
		cursor, err := collection.Find(context.TODO(), filter)
		if err != nil {
			log.Printf("%s", err)
		}
		var results []client2.Payload
		if err = cursor.All(context.TODO(), &results); err != nil {
			panic(err)
		}
		payloadMessage := Data{
			Available: len(results),
			Total:     int(estCount),
		}
		w.Header().Set("Content-Type", "application/json")
		jsonData, err := json.Marshal(payloadMessage)
		log.Printf("%s", string(jsonData))
		if err != nil {
			log.Printf("Error here")
		}
		_, err = w.Write(jsonData)
		if err != nil {
			return
		}
		//json.NewEncoder(w).Encode(payloadMessage)
	}
}

func root(w http.ResponseWriter, _ *http.Request) {
	_, err := io.WriteString(w, "Version 1.0\n")
	if err != nil {
		return
	}
}

func Start(cfg client2.Config) error {
	http.HandleFunc("/", root)
	http.HandleFunc("/servers/quick", quickServers(cfg))
	http.HandleFunc("/services/quick", quickServices(cfg))
	err := http.ListenAndServe(cfg.Server.Port, nil)
	if err != nil {
		return err
	}
	return nil
}
