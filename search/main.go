package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
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

	// check if there is ingredients filter or no
	if s.Ingredients != "" {
		splitIngredients := strings.Split(s.Ingredients, ",")

		filterTermsIngredients := map[string]interface{}{
			"terms": map[string]interface{}{
				"ingredients": splitIngredients,
			},
		}
		filter = append(filter, filterTermsIngredients)
	}

	// check if there is halal filter or no
	if s.isHalal {
		filterExistHalal := map[string]interface{}{
			"exists": map[string]interface{}{
				"field": "is_halal",
			},
		}
		filter = append(filter, filterExistHalal)
	}

	// check if there is vegetarian filter or no
	if s.isVegetarian {
		filterExistVegetarian := map[string]interface{}{
			"exists": map[string]interface{}{
				"field": "is_vegetarian",
			},
		}
		filter = append(filter, filterExistVegetarian)
	}

	// if there is no filter, just return nil
	if len(filter) == 0 {
		return nil
	}

	return filter
}

func (s searchRequest) search() searchResponse {
	var responses map[string]interface{}
	var buf bytes.Buffer

	// per page is 4
	size := 4
	from := s.Page*size - size

	// create the query
	query := map[string]interface{}{
		"from": from,
		"size": size,
		"query": map[string]interface{}{
			"function_score": map[string]interface{}{
				"query": map[string]interface{}{
					"bool": map[string]interface{}{
						"must": map[string]interface{}{
							"match": map[string]interface{}{
								"name": map[string]interface{}{
									"query": s.Keyword,
								},
							},
						},
						"filter": s.filterQuery(),
					},
				},
				"functions": []map[string]interface{}{
					{
						"filter": map[string]interface{}{
							"range": map[string]interface{}{
								"rating": map[string]interface{}{
									"gt": 4,
								},
							},
						},
						"weight": 10,
					},
				},
			},
		},
		"aggs": map[string]interface{}{
			"ingredients": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "ingredients.keyword",
				},
			},
		},
	}

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Fatalf("Error encoding query: %s", err)
	}

	fmt.Println("query: ", &buf)

	// do http call to elasticsearch
	res, err := elasticClient.Search(
		elasticClient.Search.WithContext(context.Background()),
		elasticClient.Search.WithIndex("recipes"),
		elasticClient.Search.WithBody(&buf),
		elasticClient.Search.WithTrackTotalHits(true),
	)

	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	// check if there is error or no
	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			log.Fatalf("Error parsing the response body: %s", err)
		} else {
			// Print the response status and error information.
			log.Fatalf("[%s] %s: %s",
				res.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"],
			)
		}
	}

	if err := json.NewDecoder(res.Body).Decode(&responses); err != nil {
		log.Fatalf("Error parsing the response body: %s", err)
	}
	// Print the response status, number of results, and request duration.
	log.Printf(
		"[%s] %d hits; took: %dms",
		res.Status(),
		int(responses["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
		int(responses["took"].(float64)),
	)

	// process the data
	recipes := make([]recipe, 0)

	for _, hit := range responses["hits"].(map[string]interface{})["hits"].([]interface{}) {

		source := hit.(map[string]interface{})["_source"]

		recipes = append(recipes, recipe{
			ID:           hit.(map[string]interface{})["_id"].(string),
			Name:         source.(map[string]interface{})["name"].(string),
			Ingredients:  arrInterfaceToArrString(source.(map[string]interface{})["ingredients"].([]interface{})),
			IsHalal:      interfaceToBool(source.(map[string]interface{})["is_halal"]),
			IsVegetarian: interfaceToBool(source.(map[string]interface{})["is_vegetarian"]),
			Description:  source.(map[string]interface{})["description"].(string),
			Rating:       source.(map[string]interface{})["rating"].(float64),
		})
	}

	buckets := make([]bucket, 0)

	for _, ingredientsBucket := range responses["aggregations"].(map[string]interface{})["ingredients"].(map[string]interface{})["buckets"].([]interface{}) {
		buckets = append(buckets, bucket{
			Key:      ingredientsBucket.(map[string]interface{})["key"].(string),
			DocCount: int(ingredientsBucket.(map[string]interface{})["doc_count"].(float64)),
		})
	}

	log.Println(strings.Repeat("=", 37))

	return searchResponse{
		Recipes:   recipes,
		TotalData: int(responses["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
		Filter: filter{
			IngredientBucket: buckets,
		},
	}
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
