package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Product struct {
	ProductID      int    `json:"productId"`
	Manufacturer   string `json:"manufacturer"`
	Sku            string `json:"sku"`
	Upc            string `json:"upc"`
	PricePerUnit   string `json:"pricePerUnit"`
	QuantityOnHand int    `json:"quantityOnHand"`
	ProductName    string `json:"productName"`
}

var productList []Product

func init() {
	productsJSON := `[
		{
		  "productId": 1,
		  "manufacturer": "Johns-Jenkins",
		  "sku": "p5z343vdS",
		  "upc": "939581000000",
		  "pricePerUnit": "497.45",
		  "quantityOnHand": 9703,
		  "productName": "sticky note"
		},
		{
		  "productId": 2,
		  "manufacturer": "Hessel, Schimmel and Feeney",
		  "sku": "i7v300kmx",
		  "upc": "740979000000",
		  "pricePerUnit": "282.29",
		  "quantityOnHand": 9217,
		  "productName": "leg warmers"
		},
		{
		  "productId": 3,
		  "manufacturer": "Swaniawski, Bartoletti and Bruen",
		  "sku": "q0L657ys7",
		  "upc": "111730000000",
		  "pricePerUnit": "436.26",
		  "quantityOnHand": 5905,
		  "productName": "lamp shade"
		},
		{
		  "productId": 4,
		  "manufacturer": "Runolfsdottir, Littel and Dicki",
		  "sku": "x78426lq1",
		  "upc": "93986215015",
		  "pricePerUnit": "537.90",
		  "quantityOnHand": 2642,
		  "productName": "flowers"
		}
	]`
	err := json.Unmarshal([]byte(productsJSON), &productList)
	if err != nil {
		log.Fatal(err)
	}
}

func findProductByID(productID int) (*Product, int) {
	for i, product := range productList {
		if productID == product.ProductID {
			return &productList[i], i
		}
	}
	return nil, 0
}

func productHandler(w http.ResponseWriter, r *http.Request) {
	urlPathSegments := strings.Split(r.URL.Path, "products/")
	productID, err := strconv.Atoi(urlPathSegments[len(urlPathSegments)-1])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	product, listItemIndex := findProductByID(productID)

	if listItemIndex == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		productJSON, _ := json.Marshal(product)
		w.Header().Set("Content-Type", "application/json")
		w.Write(productJSON)
	case http.MethodPut:
		var updatedProduct Product
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = json.Unmarshal(bodyBytes, &updatedProduct)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if updatedProduct.ProductID != productID {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		product = &updatedProduct
		productList[listItemIndex] = *product
		w.WriteHeader(http.StatusOK)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return

	}

}

func productsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		productsJson, err := json.Marshal(productList)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(productsJson)
	case http.MethodPost:
		var newProduct Product
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = json.Unmarshal(bodyBytes, &newProduct)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if newProduct.ProductID != 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		productList = append(productList, newProduct)
		w.WriteHeader(http.StatusCreated)
		return
	}
}

type fooHandler struct {
	Message string `json:"message"`
}

func (f *fooHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	val, _ := json.Marshal(f)
	w.Write([]byte(val))
}

func barHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("the bar"))
}

func middlewareHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("before handler; middleware start")
		start := time.Now()
		handler.ServeHTTP(w, r)
		fmt.Printf("middleware finished; %s", time.Since(start))
	})
}

func main() {
	http.Handle("/foo", &fooHandler{Message: "Go bucs"})
	http.HandleFunc("/bar", barHandler)

	productListHandler := http.HandlerFunc(productsHandler)
	productItemHandler := http.HandlerFunc(productHandler)

	http.Handle("/products", middlewareHandler(productListHandler))
	http.Handle("/products/", middlewareHandler(productItemHandler))
	http.ListenAndServe(":5000", nil)
}
