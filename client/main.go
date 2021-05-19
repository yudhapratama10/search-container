package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/lib/pq"
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
		pgConString: "postgres://postgres:postgres@localhost?sslmode=disable",
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

}

type response struct {
	Message string `json:"message,omitempty"`
}

func insert(w http.ResponseWriter, r *http.Request) {
	res := response{
		Message: "Fail to Insert",
	}

	name := r.FormValue("name")
	ingredients := r.FormValue("ingredients")
	isHalal := r.FormValue("isHalal")
	isVegetarian := r.FormValue("isVegetarian")
	description := r.FormValue("description")
	rating := r.FormValue("rating")

	splitIngredients := strings.Split(ingredients, ",")
	isHalalBool, _ := strconv.ParseBool(isHalal)
	isVegetarianBool, _ := strconv.ParseBool(isVegetarian)
	ratingFloat, _ := strconv.ParseFloat(rating, 64)

	p := recipe{
		Name:         name,
		Ingredients:  splitIngredients,
		IsHalal:      isHalalBool,
		IsVegetarian: isVegetarianBool,
		Description:  description,
		Rating:       ratingFloat,
	}

	err := p.insert()

	if err == nil {
		res.Message = "Success to Insert"
	} else {
		fmt.Println(err)
	}

	json.NewEncoder(w).Encode(res)
}

func update(w http.ResponseWriter, r *http.Request) {
	res := response{
		Message: "Fail to Update",
	}

	id := r.FormValue("id")

	name := r.FormValue("name")
	ingredients := r.FormValue("ingredients")
	isHalal := r.FormValue("isHalal")
	isVegetarian := r.FormValue("isVegetarian")
	description := r.FormValue("description")
	rating := r.FormValue("rating")

	splitIngredients := strings.Split(ingredients, ",")
	isHalalBool, _ := strconv.ParseBool(isHalal)
	isVegetarianBool, _ := strconv.ParseBool(isVegetarian)
	idInt, _ := strconv.Atoi(id)
	ratingFloat, _ := strconv.ParseFloat(rating, 64)

	p := recipe{
		ID:           idInt,
		Name:         name,
		Ingredients:  splitIngredients,
		IsHalal:      isHalalBool,
		IsVegetarian: isVegetarianBool,
		Description:  description,
		Rating:       ratingFloat,
	}

	err := p.update()

	if err == nil {
		res.Message = "Success to Update"
	} else {
		fmt.Println(err)
	}
	json.NewEncoder(w).Encode(res)
}

func delete(w http.ResponseWriter, r *http.Request) {
	res := response{
		Message: "Fail to Delete",
	}

	id := r.FormValue("id")
	idInt, _ := strconv.Atoi(id)

	p := recipe{
		ID: idInt,
	}

	err := p.delete()

	if err == nil {
		res.Message = "Success to Delete"
	}
	json.NewEncoder(w).Encode(res)
}

type recipes []recipe

func get(w http.ResponseWriter, r *http.Request) {
	recipes := recipes{}
	recipes.get()
	json.NewEncoder(w).Encode(recipes)

}

type recipe struct {
	ID           int      `json:"id,omitempty"`
	Name         string   `json:"name,omitempty"`
	Ingredients  []string `json:"ingredients,omitempty"`
	IsHalal      bool     `json:"isHalal"`
	IsVegetarian bool     `json:"isVegetarian"`
	Description  string   `json:"description,omitempty"`
	Rating       float64  `json:"rating"`
}

func (ps *recipes) get() {
	res, _ := dbCon.Query(`
	SELECT id, name, ingredients, isHalal, isVegetarian, description, rating
	FROM recipes 
	ORDER BY id desc LIMIT 5`)
	for res.Next() {
		product := recipe{}
		if err := res.Scan(&product.ID, &product.Name, pq.Array(&product.Ingredients),
			&product.IsHalal, &product.IsVegetarian, &product.Description, &product.Rating); err == nil {
			*ps = append(*ps, product)
		} else {
			fmt.Println(err)
		}
	}
}

func (p *recipe) delete() error {
	_, err := dbCon.Exec("DELETE FROM recipes where id = $1", p.ID)
	return err
}

func (p *recipe) update() error {
	_, err := dbCon.Exec(`UPDATE recipes 
	set name = $1, ingredients = $2, isHalal = $3, isVegetarian = $4, description = $5, rating = $6
	WHERE id = $7`, p.Name, pq.Array(p.Ingredients), p.IsHalal, p.IsVegetarian, p.Description, p.Rating, p.ID)
	return err
}

func (p *recipe) insert() error {
	var lastInsertID int
	err := dbCon.QueryRow(`INSERT INTO recipes(name,ingredients,isHalal,isVegetarian,description,rating) 
	values ($1,$2,$3,$4,$5,$6) 
	RETURNING id`, p.Name, pq.Array(p.Ingredients), p.IsHalal, p.IsVegetarian, p.Description, p.Rating).Scan(&lastInsertID)
	p.ID = lastInsertID
	return err
}

type message struct {
	ID     int    `json:"id,omitempty"`
	Action string `json:"action,omitempty"`
}

func (p *recipe) publishMessage(action string) {

}
