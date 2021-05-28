package main

import (
	// "bytes"
	// "context"
	"encoding/json"
	// "fmt"
	"log"
	"net/http"
	"strconv"

	// "strings"
	"text/template"

	elasticsearch "github.com/elastic/go-elasticsearch/v7"
)

var elasticClient *elasticsearch.Client

func main() {

	var err error

	elasticClient, err = elasticsearch.NewDefaultClient()
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}
	http.HandleFunc("/index", renderIndexPage)
	http.HandleFunc("/search", searchHandler)

	http.ListenAndServe(":2222", nil)

}

func renderIndexPage(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("index.html")
	_ = tmpl.Execute(w, nil)
}

type bucket struct {
	Key      string `json:"key"`
	DocCount int    `json:"doc_count"`
}
type filter struct {
	IngredientBucket []bucket `json:"ingredient_bucket"`
}
type searchResponse struct {
	TotalData int      `json:"total_data"`
	Recipes   []recipe `json:"recipes"`
	Filter    filter   `json:"filter"`
}

type recipe struct {
	ID           string   `json:"id,omitempty"`
	Name         string   `json:"name,omitempty"`
	Ingredients  []string `json:"ingredients,omitempty"`
	IsHalal      bool     `json:"isHalal"`
	IsVegetarian bool     `json:"isVegetarian"`
	Description  string   `json:"description,omitempty"`
	Rating       float64  `json:"rating"`
}

type searchRequest struct {
	Keyword      string
	Ingredients  string
	isHalal      bool
	isVegetarian bool
	Page         int
}

func (s searchRequest) filterQuery() []map[string]interface{} {

	// create the filter query variable
	filter := make([]map[string]interface{}, 0)

	// if there is no filter, just return nil
	if len(filter) == 0 {
		return nil
	}

	return filter
}

func (s searchRequest) search() searchResponse {
	return searchResponse{}
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	q := r.FormValue("q")
	ingredients := r.FormValue("ingredients")
	isHalal := r.FormValue("isHalal")
	isVegetarian := r.FormValue("isVegetarian")
	page := r.FormValue("page")

	isHalalBool, _ := strconv.ParseBool(isHalal)
	isVegetarianBool, _ := strconv.ParseBool(isVegetarian)
	pageInt, _ := strconv.Atoi(page)

	if pageInt == 0 {
		pageInt = 1
	}

	request := searchRequest{
		Keyword:      q,
		Ingredients:  ingredients,
		isHalal:      isHalalBool,
		isVegetarian: isVegetarianBool,
		Page:         pageInt,
	}

	json.NewEncoder(w).Encode(request.search())

}

func interfaceToBool(in interface{}) (out bool) {
	if in == nil {
		return false
	}

	return true
}

func arrInterfaceToArrString(in []interface{}) (out []string) {
	for _, v := range in {
		out = append(out, v.(string))
	}

	return out
}
