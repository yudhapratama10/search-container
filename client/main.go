package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/nsqio/go-nsq"
)

var producer *nsq.Producer
var dbCon *sql.DB

func main() {

	connectDB()
	createProducer()
	http.HandleFunc("/index", renderIndexPage)
	http.HandleFunc("/insert", insert)
	http.ListenAndServe(":1111", nil)

}

func connectDB() {
	var err error
	dbCon, err = sql.Open("postgres", "postgres://postgres:postgres@localhost/workshop?sslmode=disable")
	if err != nil {
		log.Fatalf("%v", err)
	}
}

func createProducer() {
	var err error
	config := nsq.NewConfig()
	producer, err = nsq.NewProducer("127.0.0.1:4150", config)

	if err != nil {
		log.Fatal(err)
	}

	defer producer.Stop()
}

func renderIndexPage(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("index.html")
	_ = tmpl.Execute(w, nil)
}

func insert(w http.ResponseWriter, r *http.Request) {

	name := r.FormValue("name")
	price := r.FormValue("price")

	priceInt, _ := strconv.Atoi(price)

	product := product{
		name:  name,
		price: priceInt,
	}

	product.insertDB()
	product.publishMessage()
}

type product struct {
	name  string
	price int
}

func (p product) insertDB() {

}

func (p product) publishMessage() {

}
