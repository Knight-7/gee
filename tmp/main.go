package main

import (
	"encoding/xml"
	"fmt"
	"net/http"
)

type User struct {
	Name string
	Age  int
}

func jsonHandler(w http.ResponseWriter, req *http.Request) {
	u := User{"Knight", 24}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(200)
	if err := xml.NewEncoder(w).Encode(u); err != nil {
		panic(err)
	}
}

func tmpHandler(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("hello world"))
	fmt.Println(w.Header().Get("Content-Type"))
}

func main() {
	http.HandleFunc("/json", jsonHandler)
	http.HandleFunc("/tmp", tmpHandler)
	http.ListenAndServe(":2020", nil)
}
