package main

import (
	"net/http"
	"text/template"
)

func main() {
	http.HandleFunc("/index", renderIndexPage)
	http.ListenAndServe(":2222", nil)

}

func renderIndexPage(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("index.html")
	_ = tmpl.Execute(w, nil)
}

type searchResponse struct {
	TotalData string   `json:"total_data,omitempty"`
	Recipes   []recipe `json:"recipes,omitempty"`
}

type recipe struct {
	ID           int      `json:"id,omitempty"`
	Name         string   `json:"name,omitempty"`
	Ingredients  []string `json:"ingredients,omitempty"`
	IsHalal      bool     `json:"isHalal"`
	IsVegetarian bool     `json:"isVegetarian"`
	Description  string   `json:"description,omitempty"`
	Rating       int      `json:"rating"`
}

func searchHandler(w http.ResponseWriter, r *http.Request) {

}

type filterResponse struct {
}

func filterHandler(w http.ResponseWriter, r *http.Request) {

}
