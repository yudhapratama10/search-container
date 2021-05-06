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

	_ "github.com/lib/pq"
)

func main() {

	dbConn, err := sql.Open("postgres", "postgres://postgres@localhost:5432/?sslmode=disable")
	if err != nil {
		panic(err)
	} else if err := dbConn.Ping(); err != nil {
		panic(err)
	}

	rows, err := dbConn.QueryContext(context.Background(), "SELECT id, name, price FROM product")
	if err != nil {
		panic(err)
	}

	type product struct {
		ID    int    `json:"-"` // ID on ES are mapped into HTTP URL instead of body
		Name  string `json:"name"`
		Price int    `json:"price"`
	}
	var products []product

	for rows.Next() {
		var product product
		if err := rows.Scan(&product.ID, &product.Name, &product.Price); err != nil {
			panic(err)
		}
		products = append(products, product)
	}

	body := bytes.NewBuffer(nil)
	for _, product := range products {

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
		bulkHead.Index.Index = "product"
		bulkHead.Index.ID = product.ID

		{
			// Head
			b, _ := json.Marshal(bulkHead)
			body.Write(b)
			body.WriteByte('\n')
		}
		{
			// Tail
			b, _ := json.Marshal(product)
			body.Write(b)
			body.WriteByte('\n')
		}
	}
	req, _ := http.NewRequest(http.MethodPost, "http://localhost:9200/_bulk?pretty", body)
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	log.Println("Elastic response:")
	io.Copy(os.Stdout, res.Body)
}
