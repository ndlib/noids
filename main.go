package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/dbrower/noids/server"
	"github.com/gorilla/pat"
	flag "github.com/ogier/pflag"
)

var (
	logfile string
	logw    *os.File
)

func main() {
	var port int
	var storageDir string

	flag.IntVarP(&port, "port", "p", 8080, "port to run on")
	flag.StringVarP(&logfile, "log", "l", "", "name of log file")
	flag.StringVarP(&storageDir, "storage", "s", "pools", "directory to save noid information")

	flag.Parse()

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	if logfile != "" {
		f, err := os.OpenFile(logfile, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			log.Fatal(err)
		}
		logw = f
		log.SetOutput(f)
	}

	_ = server.StartSaver(storageDir)

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
