package main

import (
	"log"
	"net/http"

	"github.com/dbrower/noids/server"
	"github.com/gorilla/pat"
)

func main() {
	r := pat.New()
	r.Get("/pools/{poolname}", server.PoolShowHandler)
	r.Put("/pools/{poolname}/open", server.PoolOpenHandler)
	r.Put("/pools/{poolname}/close", server.PoolCloseHandler)
	r.Post("/pools/{poolname}/mint", server.MintHandler)
	// r.Get("/stats", StatsHandler)
	r.Get("/pools", server.PoolsHandler)
	r.Post("/pools", server.NewPoolHandler)

	http.Handle("/", r)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
