package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/gocolly/colly"
)

type customerDto struct {
	Name          string `json:"name"`
	FavoriteSnack string `json:"favoriteSnack"`
	TotalSnacks   int    `json:"totalSnacks"`
}

type customer struct {
	Name   string
	Snacks map[string]int
}

var customersMap = make(map[string]*customer)
var customerDtos = []customerDto{}
var customersData = []customer{}

func main() {
	collector := colly.NewCollector()

	collector.OnHTML("#top\\.customers tbody tr", accumulateCustomers)

	collector.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	collector.OnError(func(_ *colly.Response, err error) {
		fmt.Println("Error:", err)
	})

	collector.Visit("https://candystore.zimpler.net")

	for _, c := range customersData {
		favoriteAmount := 0
		var favoriteSnack string
		for key, snackAmount := range c.Snacks {
			if snackAmount > favoriteAmount {
				favoriteAmount = snackAmount
				favoriteSnack = key
			}
		}
		customerDto := customerDto{
			Name:          c.Name,
			FavoriteSnack: favoriteSnack,
			TotalSnacks:   favoriteAmount,
		}

		customerDtos = append(customerDtos, customerDto)
	}

	b, err := json.Marshal(customerDtos)
	if err != nil {
		log.Fatalf("Marshal error: %v", err)
	}

	var result bytes.Buffer
	json.Indent(&result, b, "", "   ")
	fmt.Println(result.String())
}

func accumulateCustomers(row *colly.HTMLElement) {
	name := row.ChildText("td:nth-child(1)")
	snack := row.ChildText("td:nth-child(2)")
	amountString := row.ChildText("td:nth-child(3)")

	amount, err := strconv.Atoi(amountString)

	if err != nil {
		log.Fatalf("Parsing snack amount failed: %v", err)
	}

	if _, contains := customersMap[name]; !contains {
		customer := customer{
			Name:   name,
			Snacks: make(map[string]int),
		}
		customer.Snacks[snack] = amount

		customersData = append(customersData, customer)
		customersMap[name] = &customersData[len(customersData)-1]
		return
	}

	if _, contains := customersMap[name].Snacks[snack]; !contains {
		customersMap[name].Snacks[snack] = amount
		return
	}

	customersMap[name].Snacks[snack] += amount
}
