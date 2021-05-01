package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	_ "github.com/lib/pq"

	"github.com/nsqio/go-nsq"
)

var producer *nsq.Producer
var dbCon *sql.DB

type conSetting struct {
	nsqAddr     string
	pgConString string
}

func main() {
	defer producer.Stop()
	c := conSetting{
		nsqAddr:     "127.0.0.1:4150",
		pgConString: "postgres://postgres:postgres@localhost/workshop?sslmode=disable",
	}

	c.connectDB()
	c.createProducer()
	http.HandleFunc("/index", renderIndexPage)
	http.HandleFunc("/insert", insert)
	http.HandleFunc("/update", update)
	http.HandleFunc("/delete", delete)
	http.HandleFunc("/get", get)

	http.ListenAndServe(":1111", nil)

}

func renderIndexPage(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("index.html")
	_ = tmpl.Execute(w, nil)
}

func (c conSetting) connectDB() {
	var err error
	dbCon, err = sql.Open("postgres", c.pgConString)
	if err != nil {
		log.Fatalf("%v", err)
	}
}

func (c conSetting) createProducer() {
	var err error
	config := nsq.NewConfig()
	producer, err = nsq.NewProducer(c.nsqAddr, config)

	if err != nil {
		log.Fatal(err)
	}

}

type response struct {
	Message string `json:"message,omitempty"`
}

func insert(w http.ResponseWriter, r *http.Request) {
	res := response{
		Message: "Fail to Insert",
	}

	name := r.FormValue("name")
	price := r.FormValue("price")

	priceInt, _ := strconv.Atoi(price)
	p := product{
		Name:  name,
		Price: priceInt,
	}

	err := p.insert()

	if err == nil {
		p.publishMessage("insert")
		res.Message = "Success to Insert"
	}
	json.NewEncoder(w).Encode(res)
}

func update(w http.ResponseWriter, r *http.Request) {
	res := response{
		Message: "Fail to Update",
	}

	id := r.FormValue("id")
	name := r.FormValue("name")
	price := r.FormValue("price")

	priceInt, _ := strconv.Atoi(price)
	IDInt, _ := strconv.Atoi(id)

	p := product{
		ID:    IDInt,
		Name:  name,
		Price: priceInt,
	}

	err := p.update()

	if err == nil {
		p.publishMessage("update")
		res.Message = "Success to Update"
	}
	json.NewEncoder(w).Encode(res)
}

func delete(w http.ResponseWriter, r *http.Request) {
	res := response{
		Message: "Fail to Delete",
	}

	id := r.FormValue("id")
	IDInt, _ := strconv.Atoi(id)

	p := product{
		ID: IDInt,
	}

	err := p.delete()

	if err == nil {
		p.publishMessage("delete")
		res.Message = "Success to Update"
	}
	json.NewEncoder(w).Encode(res)
}

type products []product

func get(w http.ResponseWriter, r *http.Request) {
	products := products{}
	products.get()
	json.NewEncoder(w).Encode(products)

}

type product struct {
	ID    int    `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	Price int    `json:"price,omitempty"`
}

func (ps *products) get() {
	res, _ := dbCon.Query("SELECT id, name, price FROM products")

	for res.Next() {
		product := product{}
		if err := res.Scan(&product.ID, &product.Name, &product.Price); err == nil {
			fmt.Println(product)
			*ps = append(*ps, product)
		}
	}
}

func (p *product) delete() error {
	_, err := dbCon.Exec("DELETE FROM products where id = $1", p.ID)
	return err
}

func (p *product) update() error {
	_, err := dbCon.Exec("UPDATE products set name = $1, price = $2 WHERE id = $3", p.Name, p.Price, p.ID)
	return err
}

func (p *product) insert() error {
	var lastInsertID int
	err := dbCon.QueryRow("INSERT INTO products(name,price) values ($1,$2) RETURNING id", p.Name, p.Price).Scan(&lastInsertID)
	p.ID = lastInsertID
	return err
}

type message struct {
	ID     int    `json:"id,omitempty"`
	Action string `json:"action,omitempty"`
}

func (p *product) publishMessage(action string) {
	msg := message{
		ID:     p.ID,
		Action: action,
	}
	payload, err := json.Marshal(msg)
	if err != nil {
		log.Println(err)
	}
	err = producer.Publish("product", payload)
	if err != nil {
		log.Println(err)
	}
}
