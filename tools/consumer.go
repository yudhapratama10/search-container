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

	consumer, err := nsq.NewConsumer("product_index", "search_service", nsq.NewConfig())
	if err != nil {
		panic(err)
	}

	consumer.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {

		var product struct {
			ID int `json:"id"`
		}
		if err := json.Unmarshal(message.Body, &product); err != nil {
			message.Finish()
			return nil // invalid message are dropped for simplicity
		} else if product.ID == 0 {
			message.Finish()
			return nil
		}

		req, _ := http.NewRequest(
			http.MethodPut,
			fmt.Sprintf("http://localhost:9200/product/_doc/%d?pretty", product.ID),
			bytes.NewBuffer(message.Body),
		)
		req.Header.Add("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		// Message are passed to ES succesfully, sending FIN command
		message.Finish()

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

	if err := consumer.ConnectToNSQLookupd("http://localhost:4161"); err != nil {
		panic(err)
	}

	go func(consumer *nsq.Consumer) {

		// Graceful handling
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
