package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"github.com/lib/pq"
	"github.com/nsqio/go-nsq"
)

func main() {

	isFullImport := flag.Bool("fullimport", false, "full import mode")
	flag.Parse()

	if *isFullImport {
		importer()
	} else {
		consumer()
	}
}

func consumer() {

	config := nsq.NewConfig()
	// Create new NSQ consumer, for consuming new message
	consumer, err := nsq.NewConsumer("recipes", "search_service", config)
	if err != nil {
		panic(err)
	}

	// Adds a handler, basically what we want to do everytime we consume a message
	consumer.AddHandler(nsq.HandlerFunc(func(m *nsq.Message) error {

		// Parse message body into a struct
		var message struct {
			ID     int    `json:"id"`
			Action string `json:"action"`
		}
		if err := json.Unmarshal(m.Body, &message); err != nil {
			return nil // invalid message are dropped for simplicity
		} else if message.ID == 0 {
			return nil
		} else if message.Action == "" {
			return nil
		}

		// Recipe data body representation
		var recipe struct {
			Name         string   `json:"name"`
			Ingredients  []string `json:"ingredients"`
			IsHalal      bool     `json:"is_halal,omitempty"`
			IsVegetarian bool     `json:"is_vegetarian,omitempty"`
			Description  string   `json:"description"`
			Rating       float64  `json:"rating"`
		}

		// Get detailed data from SQL/database
		dbConn, err := sql.Open("postgres", "postgres://postgres@localhost:5432/?sslmode=disable")
		if err != nil {
			fmt.Println("NSQ-ES consumer error: Database:", err)
			return nil
		} else if err := dbConn.Ping(); err != nil {
			fmt.Println("NSQ-ES consumer error: Database:", err)
			return nil
		}
		if err := dbConn.QueryRowContext(context.Background(), `
		SELECT
			name,
			ingredients,
			isHalal,
			isVegetarian,
			description,
			rating
		FROM
			recipes
		WHERE
			id = $1
		`, message.ID).Scan(
			&recipe.Name,
			pq.Array(&recipe.Ingredients),
			&recipe.IsHalal,
			&recipe.IsVegetarian,
			&recipe.Description,
			&recipe.Rating,
		); err != nil {
			fmt.Println("NSQ-ES consumer error: Database:", err, "ID:", message.ID)
			return nil
		}

		// Construct an HTTP request using the struct data
		var req *http.Request
		if message.Action == "insert" || message.Action == "update" {
			marshal, _ := json.Marshal(recipe)
			req, _ = http.NewRequest(
				http.MethodPut,
				fmt.Sprintf("http://localhost:9200/recipes/_doc/%d?pretty", message.ID),
				bytes.NewBuffer(marshal),
			)
		} else if message.Action == "delete" {
			req, _ = http.NewRequest(
				http.MethodDelete,
				fmt.Sprintf("http://localhost:9200/recipes/_doc/%d?pretty", message.ID),
				nil,
			)
		} else {
			fmt.Println("NSQ-ES consumer error: Request: Invalid `Action` type", message.Action)
			return nil
		}
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
				<-c
				log.Println("Stopping consumer...")
				consumer.Stop()
			}()
		}
	}(consumer)

	<-consumer.StopChan
}

func importer() {
	fmt.Println(strings.Repeat("=", 30))
	fmt.Println("FullImport Started")
	fmt.Println(strings.Repeat("=", 30))

	// Connect to database
	dbConn, err := sql.Open("postgres", "postgres://postgres@localhost:5432/?sslmode=disable")
	if err != nil {
		panic(err)
	} else if err := dbConn.Ping(); err != nil {
		panic(err)
	}

	// Do SQL query to get the data
	rows, err := dbConn.QueryContext(context.Background(), `
	SELECT
		id,
		name,
		ingredients,
		isHalal,
		isVegetarian,
		description,
		rating
	FROM
		recipes`)
	if err != nil {
		panic(err)
	}

	// Define data schema
	type recipe struct {
		ID           int      `json:"-"` // ID on ES are mapped into HTTP URL instead of body
		Name         string   `json:"name"`
		Ingredients  []string `json:"ingredients"`
		IsHalal      bool     `json:"is_halal,omitempty"`
		IsVegetarian bool     `json:"is_vegetarian,omitempty"`
		Description  string   `json:"description"`
		Rating       float64  `json:"rating"`
	}

	// Using the result from query, map it to the schema
	var recipes []recipe
	for rows.Next() {
		var recipe recipe
		if err := rows.Scan(
			&recipe.ID,
			&recipe.Name,
			pq.Array(&recipe.Ingredients),
			&recipe.IsHalal,
			&recipe.IsVegetarian,
			&recipe.Description,
			&recipe.Rating,
		); err != nil {
			panic(err)
		}
		recipes = append(recipes, recipe)
	}
	rows.Close()

	// Parse the data into HTTP request body bytes
	body := bytes.NewBuffer(nil)
	for _, recipe := range recipes {

		// Elastic bulk insert uses head-tail (for index) format.
		// Ex:
		// POST _bulk
		// { "index" : { "_index" : "test", "_id" : "1" } }
		// { "field1" : "value1" }
		// { "create" : { "_index" : "test", "_id" : "3" } }
		// { "field1" : "value3" }
		// { "update" : {"_id" : "1", "_index" : "test"} }
		// { "doc" : {"field2" : "value2"} }

		var bulkHead struct {
			Index struct {
				Index string `json:"_index"`
				ID    int    `json:"_id"`
			} `json:"index"`
		}
		bulkHead.Index.Index = "recipes"
		bulkHead.Index.ID = recipe.ID

		{
			// Head
			b, _ := json.Marshal(bulkHead)
			body.Write(b)
			body.WriteByte('\n')
		}
		{
			// Tail
			b, _ := json.Marshal(recipe)
			body.Write(b)
			body.WriteByte('\n')
		}
	}
	req, _ := http.NewRequest(http.MethodPost, "http://localhost:9200/_bulk?pretty", body)
	req.Header.Add("Content-Type", "application/json")

	// With that HTTP request, POST the data into Elasticsearch (index/insert data)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	log.Println("Elastic response:")
	io.Copy(os.Stdout, res.Body)

	fmt.Println(strings.Repeat("=", 30))
	fmt.Println("FullImport Done")
	fmt.Println(strings.Repeat("=", 30))

}
