package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	//server startup in GCP Cloudrun

	/*
		a serveMux is a http request multiplixer
		it matches the url of each request against registered patterns
		and calls appropriate handler based on the incoming url
	*/

	mux := http.NewServeMux()
	mux.HandleFunc("POST /chat", ChatHandler)
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	log.Printf("server started in port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux)) // log error if crash occurs

}
