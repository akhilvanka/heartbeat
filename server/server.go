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

type DataServer struct {
	Available int `json:"available"`
	Total     int `json:"total"`
}

func quickServers(w http.ResponseWriter, _ *http.Request) {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI("").SetServerAPIOptions(serverAPI)

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

	filter := bson.D{{"upload_time", bson.D{{"$lt", time.Now().Unix() - 300000}}}}
	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		log.Printf("%s", err)
	}
	var results []client2.Payload
	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}
	payloadMessage := DataServer{
		Available: len(results),
		Total:     int(estCount),
	}
	w.Header().Set("Content-Type", "application/json")
	log.Printf("Got here")
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

func root(w http.ResponseWriter, _ *http.Request) {
	_, err := io.WriteString(w, "Version 1.0\n")
	if err != nil {
		return
	}
}

func Start() error {
	http.HandleFunc("/", root)
	http.HandleFunc("/servers/quick", quickServers)
	err := http.ListenAndServe(":1212", nil)
	if err != nil {
		return err
	}
	return nil
}
