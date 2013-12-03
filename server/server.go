package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

// Implements the Noid server API

func PoolsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.RequestURI)
	names := AllPools()
	enc := json.NewEncoder(w)
	enc.Encode(names)
}

func NewPoolHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.RequestURI)
	name := r.FormValue("name")
	template := r.FormValue("template")
	if name == "" || template == "" {
		http.Error(w, "missing arguments", 400)
		return
	}
	pi, err := AddPool(name, template)
	if err != nil {
		if err == NameExists {
			http.Error(w, "name already exists", 409)
		} else {
			http.Error(w, err.Error(), 400)
		}
		return
	}
	w.WriteHeader(201)
	enc := json.NewEncoder(w)
	enc.Encode(pi)
}

func PoolShowHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.RequestURI)
	name := r.FormValue(":poolname")
	pi, err := GetPool(name)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	enc := json.NewEncoder(w)
	enc.Encode(pi)
}

func PoolOpenHandler(w http.ResponseWriter, r *http.Request) {
	handleOpenClose(w, r, false)
}

func PoolCloseHandler(w http.ResponseWriter, r *http.Request) {
	handleOpenClose(w, r, true)
}

func handleOpenClose(w http.ResponseWriter, r *http.Request, makeClosed bool) {
	log.Println(r.RequestURI)
	name := r.FormValue(":poolname")
	pi, err := SetPoolState(name, makeClosed)
	if err != nil {
		http.Error(w, err.Error(), 403)
		return
	}
	enc := json.NewEncoder(w)
	enc.Encode(pi)
}

func MintHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.RequestURI)
	var count int = 1
	var err error

	name := r.FormValue(":poolname")
	n := r.FormValue("n")

	if n != "" {
		count, err = strconv.Atoi(n)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		if count <= 0 || count > 1000 {
			http.Error(w, "count is out of range", 400)
			return
		}
	}

	ids, err := PoolMint(name, count)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	enc := json.NewEncoder(w)
	enc.Encode(ids)
}

func AdvancePastHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.RequestURI)

	name := r.FormValue(":poolname")
	id := r.FormValue("id")

	if id == "" {
		http.Error(w, "id parameter is required", 400)
		return
	}

	pi, err := PoolAdvancePast(name, id)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	enc := json.NewEncoder(w)
	enc.Encode(pi)
}
