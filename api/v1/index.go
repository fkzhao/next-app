package handler

import (
	"fmt"
	"net/http"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	_, err := fmt.Fprintf(w, "<h1>Hello from Go!</h1>")
	if err != nil {
		return
	}
}
