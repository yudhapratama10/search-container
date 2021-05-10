package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/lib/pq"
)

func main() {

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
		IsHalal      bool     `json:"is_halal"`
		IsVegetarian bool     `json:"is_vegetarian"`
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
}
