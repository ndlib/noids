package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/pat"
)

var (
	pools *poolGroup = NewPoolGroup()
)

// Implements the Noid server API

func PoolsHandler(w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	writeJSON(w, pools.AllPools())
}

func NewPoolHandler(w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	name := r.FormValue("name")
	template := r.FormValue("template")
	if name == "" || template == "" {
		http.Error(w, "missing arguments", 400)
		return
	}
	pi, err := pools.AddPool(name, template)
	if err != nil {
		if err == NameExists {
			http.Error(w, "name already exists", 409)
		} else {
			log.Println("Error:", err)
			http.Error(w, err.Error(), 400)
		}
		return
	}
	writeJSONStatus(w, pi, 201)
}

func PoolShowHandler(w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	name := r.FormValue(":poolname")
	pi, err := pools.GetPool(name)
	if err != nil {
		// most likely the error is that the pool name doesn't exist
		log.Println("Error:", err)
		http.Error(w, err.Error(), 404)
		return
	}
	writeJSON(w, pi)
}

func PoolOpenHandler(w http.ResponseWriter, r *http.Request) {
	handleOpenClose(w, r, false)
}

func PoolCloseHandler(w http.ResponseWriter, r *http.Request) {
	handleOpenClose(w, r, true)
}

func handleOpenClose(w http.ResponseWriter, r *http.Request, makeClosed bool) {
	logRequest(r)
	name := r.FormValue(":poolname")
	pi, err := pools.SetPoolState(name, makeClosed)
	if err != nil {
		log.Println("Error:", err)
		http.Error(w, err.Error(), 403)
		return
	}
	writeJSON(w, pi)
}

func MintHandler(w http.ResponseWriter, r *http.Request) {
	logRequest(r)
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

	ids, err := pools.PoolMint(name, count)
	if err != nil {
		log.Println("Error:", err)
		http.Error(w, err.Error(), 400)
		return
	}
	log.Println("Minted", ids)
	writeJSON(w, ids)
}

func AdvancePastHandler(w http.ResponseWriter, r *http.Request) {
	logRequest(r)

	name := r.FormValue(":poolname")
	id := r.FormValue("id")

	if id == "" {
		http.Error(w, "id parameter is required", 400)
		return
	}

	pi, err := pools.PoolAdvancePast(name, id)
	if err != nil {
		log.Println("Error:", err)
		http.Error(w, err.Error(), 400)
		return
	}
	writeJSON(w, pi)
}

// For now the stats provided are meager
type stats struct {
	Version string
}

func StatsHandler(w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	s := stats{
		Version: version,
	}
	writeJSON(w, s)
}

func logRequest(r *http.Request) {
	log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.RequestURI)
}

// use the standard library json Encoder to serialize `value` to `w`.
// Tries to set the content type to application/json.
// Returns status code 200. Use writeJSONStatus() to set a status code.
func writeJSON(w http.ResponseWriter, value interface{}) {
	writeJSONStatus(w, value, 200)
}

func writeJSONStatus(w http.ResponseWriter, value interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.Encode(value)
}

// SetupHandlers adds the routes to the global http route table.
// If s is not nil, then s will be set as the default store
// and all the pools will be loaded from s.
func SetupHandlers(s PoolStore) {
	if s != nil {
		DefaultStore = s
		err := pools.LoadPoolsFromStore(s)
		if err != nil {
			log.Fatal(err)
		}
	}
	r := pat.New()
	r.Get("/pools/{poolname}", PoolShowHandler)
	r.Put("/pools/{poolname}/open", PoolOpenHandler)
	r.Put("/pools/{poolname}/close", PoolCloseHandler)
	r.Post("/pools/{poolname}/mint", MintHandler)
	r.Post("/pools/{poolname}/advancePast", AdvancePastHandler)
	r.Get("/stats", StatsHandler)
	r.Get("/pools", PoolsHandler)
	r.Post("/pools", NewPoolHandler)

	http.Handle("/", r)
}
