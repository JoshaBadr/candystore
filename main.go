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

// Creates a collector to scrape candystore.zimpler.net customers
// and then prints out the favorite snack and total amount of each unique customer.
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

	// Looks for the highest total amount of snacks eaten, for each customer,
	// then creates a customer DTO to be printed out in indented JSON format.
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

// Accumulates the customers inside customer table with id=top.customers.
// If a customer doesn't exist in the map it adds that customer to the customer data slice
// and passes the adress of that customer to the map key as the value.
// Once the customer is added any new snack will be added with the amount.
// If all of the above exists as data then it will accumulate the amount to each
// corresponding snack.
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
