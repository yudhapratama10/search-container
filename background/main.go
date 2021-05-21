package main

import (
	// "bytes"
	// "context"
	// "database/sql"
	// "encoding/json"
	"flag"
	// "fmt"
	// "io"
	// "log"
	// "net/http"
	// "os"
	// "os/signal"
	// "strings"
	// "github.com/lib/pq"
	// "github.com/nsqio/go-nsq"
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

	// Create new NSQ consumer, for consuming new message

	// Adds a handler, basically what we want to do everytime we consume a message
	{
		// Parse message body into a struct

		// Recipe data body representation

		// Construct an HTTP request using the struct data

		// Index (insert) the data to Elasticsearch via PUT request

		// Might be optional: see if Elasticsearch returns any error responses.
		// If there's any, just log the response
	}

	// Using current consumer & handler, connect to a producer

	// Resiliency: graceful handling, stop the consumer on SIGINT
}

func importer() {

	// Connect to database

	// Do SQL query to get the data

	// Define data schema

	// Using the result from query, map it to the schema

	// Parse the data into HTTP request body bytes

	// With that HTTP request, POST the data into Elasticsearch (index/insert data)
}
