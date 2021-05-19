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

// fill the code for consumer here
func consumer() {

}

// fill the code for importer here
func importer() {

}
