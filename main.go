package main

import (
	"fmt"
	"log"
	"net/http"
	"next-app/core"
)

func main() {
	port := ":9000"
	log.Println(fmt.Sprintf("sever run on [%v]", 9000))
	err := http.ListenAndServe(port, http.HandlerFunc(DoHandle))
	if err != nil {
		log.Fatal(err)
		return
	}
}

func DoHandle(w http.ResponseWriter, r *http.Request) {
	core.DoProxy(w, r)
}
