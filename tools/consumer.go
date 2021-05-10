package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/nsqio/go-nsq"
)

func main() {

	// Create new NSQ consumer, for consuming new message
	consumer, err := nsq.NewConsumer("recipes_index", "search_service", nsq.NewConfig())
	if err != nil {
		panic(err)
	}

	// Adds a handler, basically what we want to do everytime we consume a message
	consumer.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {

		// Parse message body into a struct
		var recipe struct {
			ID int `json:"id"`
		}
		if err := json.Unmarshal(message.Body, &recipe); err != nil {
			message.Finish()
			return nil // invalid message are dropped for simplicity
		} else if recipe.ID == 0 {
			message.Finish()
			return nil
		}

		// Construct an HTTP request using the struct data
		req, _ := http.NewRequest(
			http.MethodPut,
			fmt.Sprintf("http://localhost:9200/recipes/_doc/%d?pretty", recipe.ID),
			bytes.NewBuffer(message.Body),
		)
		req.Header.Add("Content-Type", "application/json")

		// Index (insert) the data to Elasticsearch via PUT request
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		// Might be optional: see if Elasticsearch returns any error responses.
		// If there's any, just log the response
		var response struct {
			Error struct {
				RootCause []struct {
					Type   string `json:"type"`
					Reason string `json:"reason"`
				} `json:"root_cause"`
			} `json:"error"`
		}
		if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
			fmt.Println("NSQ-ES consumer error: Decode:", err)
			return nil
		}
		if len(response.Error.RootCause) != 0 {
			fmt.Println("NSQ-ES consumer error: Request:")
			for _, err := range response.Error.RootCause {
				fmt.Printf("%+v\n", err)
			}
		}

		return nil
	}))

	// Using current consumer & handler, connect to a producer
	if err := consumer.ConnectToNSQD("localhost:4150"); err != nil {
		panic(err)
	}

	// Resiliency: graceful handling, stop the consumer on SIGINT
	go func(consumer *nsq.Consumer) {
		{
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt)
			go func() {
				select {
				case <-c:
					log.Println("Stopping consumer...")
					consumer.Stop()
				}
			}()
		}
	}(consumer)

	<-consumer.StopChan
}
