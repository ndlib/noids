package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"code.google.com/p/gcfg"
	_ "code.google.com/p/go-sqlite/go1/sqlite3"
	_ "github.com/go-sql-driver/mysql"

	"github.com/dbrower/noids/server"
)

type Reopener interface {
	Reopen()
}

type loginfo struct {
	name string
	f    *os.File
}

func NewReopener(filename string) *loginfo {
	return &loginfo{name: filename}
}

func (li *loginfo) Reopen() {
	if li.name == "" {
		return
	}
	if li.f != nil {
		log.Println("Reopening Log files")
	}
	newf, err := os.OpenFile(li.name, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(newf)
	if li.f != nil {
		li.f.Close()
	}
	li.f = newf
}

func signalHandler(sig <-chan os.Signal, logw Reopener) {
	for s := range sig {
		log.Println("Received signal", s)
		switch s {
		case syscall.SIGUSR1:
			logw.Reopen()
		case syscall.SIGINT, syscall.SIGTERM:
			log.Println("Exiting")
			os.Exit(1)
		}
	}
}

type Config struct {
	General struct {
		Port string
	}
	Mysql struct {
		User     string
		Password string
		Host     string
		Port     string
		Database string
	}
}

func main() {
	var (
		port          string
		storageDir    string
		logfilename   string
		logw          Reopener
		sqliteFile    string
		mysqlLocation string
		showVersion   bool
		configFile    string
		config        Config
	)

	flag.StringVar(&port, "port", "13001", "port to run on")
	flag.StringVar(&logfilename, "log", "", "name of log file")
	flag.StringVar(&storageDir, "storage", "", "directory to save noid information")
	flag.StringVar(&sqliteFile, "sqlite", "", "sqlite database file to save noid information")
	flag.StringVar(&mysqlLocation, "mysql", "", "MySQL database to save noid information")
	flag.BoolVar(&showVersion, "version", false, "Display binary version")
	flag.StringVar(&configFile, "config", "", "config file to use")

	flag.Parse()

	if showVersion {
		fmt.Printf("noids version %s\n", version)
		return
	}

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	logw = NewReopener(logfilename)
	logw.Reopen()
	log.Println("-----Starting Server")

	if configFile != "" {
		log.Printf("Reading config file %s", configFile)
		err := gcfg.ReadFileInto(&config, configFile)
		if err != nil {
			log.Fatal(err)
		} else {
			// config file overrides command line
			if config.General.Port != "" {
				port = config.General.Port
			}
			if config.Mysql.Database != "" {
				mysqlLocation = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
					config.Mysql.User,
					config.Mysql.Password,
					config.Mysql.Host,
					config.Mysql.Port,
					config.Mysql.Database)
			}
		}
	}

	sig := make(chan os.Signal, 5)
	signal.Notify(sig)
	go signalHandler(sig, logw)

	var (
		saver server.PoolSaver
		db    *sql.DB
		err   error
	)
	switch {
	case storageDir != "":
		log.Println("Pool storage is directory", storageDir)
		saver = server.NewJsonFileSaver(storageDir)
	case sqliteFile != "":
		log.Println("Pool storage is sqlite3 database", sqliteFile)
		db, err = sql.Open("sqlite3", sqliteFile)
	case mysqlLocation != "":
		log.Println("Pool storage is MySQL database", mysqlLocation)
		db, err = sql.Open("mysql", mysqlLocation)
	}
	if err != nil {
		log.Fatalf("Error opening database: %s", err.Error())
	}
	if db != nil {
		saver = server.NewDbFileSaver(db)
	}
	server.SetupHandlers(saver)
	log.Println("Listening on port", port)
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
