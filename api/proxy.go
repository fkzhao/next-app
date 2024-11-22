package handler

import (
	"net/http"
	"next-app/core"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	core.DoProxy(w, r)
}
