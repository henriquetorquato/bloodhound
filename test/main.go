package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/login", getLoginPage)
	http.HandleFunc("/logout", getLogoutPage)
	http.HandleFunc("/search", getSearchPage)

	http.HandleFunc("/products", getProducts)

	err := http.ListenAndServe(":5555", nil)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}

func getLoginPage(w http.ResponseWriter, r *http.Request) {
	content, _ := os.ReadFile("./templates/login.html")
	io.Writer.Write(w, content)
}

func getLogoutPage(w http.ResponseWriter, r *http.Request) {
	content, _ := os.ReadFile("./templates/logout.html")
	io.Writer.Write(w, content)
}

func getSearchPage(w http.ResponseWriter, r *http.Request) {
	content, _ := os.ReadFile("./templates/search.html")
	io.Writer.Write(w, content)
}

func getProducts(w http.ResponseWriter, r *http.Request) {
	data := `[
		{
			"id": 1,
			"name": "potato"
		},
		{
			"id": 2,
			"name": "milk"
		}
	]`

	io.WriteString(w, data)
}
