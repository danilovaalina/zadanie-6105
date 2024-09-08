package main

import (
	"net/http"

	"zadanie-6105/internal/api"
)

func main() {
	a := api.New()

	http.ListenAndServe(":8080", a)
}
