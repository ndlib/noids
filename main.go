package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/dbrower/noids/server"
	"github.com/gorilla/pat"
	flag "github.com/ogier/pflag"
)

var (
	logfile string
	logw    *os.File
)

func signalHandler(sig <-chan os.Signal) {
	for s := range sig {
		log.Println("Got", s)
		switch s {
		case syscall.SIGUSR1:
			rotateLogFile()
		}
	}
}

func rotateLogFile() {
	if logfile == "" {
		return
	}
	if logw != nil {
		log.Println("Reopening Log files")
	}
	f, err := os.OpenFile(logfile, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(f)
	if logw != nil {
		logw.Close()
	}
	logw = f
}

func main() {
	var port string
	var storageDir string

	flag.StringVarP(&port, "port", "p", "8080", "port to run on")
	flag.StringVarP(&logfile, "log", "l", "", "name of log file")
	flag.StringVarP(&storageDir, "storage", "s", "", "directory to save noid information")

	flag.Parse()

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	rotateLogFile() // opens the log file the first time
	log.Println("-----Starting Server")

	if storageDir != "" {
		server.StartSaver(server.NewJsonFileSaver(storageDir))
	}
	sig := make(chan os.Signal, 5)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGUSR1, syscall.SIGUSR2)
	go signalHandler(sig)

	r := pat.New()
	r.Get("/pools/{poolname}", server.PoolShowHandler)
	r.Put("/pools/{poolname}/open", server.PoolOpenHandler)
	r.Put("/pools/{poolname}/close", server.PoolCloseHandler)
	r.Post("/pools/{poolname}/mint", server.MintHandler)
	r.Post("/pools/{poolname}/advancePast", server.AdvancePastHandler)
	// r.Get("/stats", StatsHandler)
	r.Get("/pools", server.PoolsHandler)
	r.Post("/pools", server.NewPoolHandler)

	http.Handle("/", r)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
