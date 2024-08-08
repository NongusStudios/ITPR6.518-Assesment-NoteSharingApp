package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/v5/stdlib" //use pgx in database/sql mode
)

// PostgreSQl configuration if not passed as env variables
const (
	host     = "localhost" //127.0.0.1
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "ESD"
)

var (
	err  error
	wait time.Duration
)

type App struct {
	Router   *mux.Router
	db       *sql.DB
	bindport string
	username string
	role     string
}

func (a *App) Initialize() {
	a.bindport = "80"

	//check if a different bind port was passed from the CLI
	//os.Setenv("PORT", "8080")
	tempport := os.Getenv("PORT")
	if tempport != "" {
		a.bindport = tempport
	}

	if len(os.Args) > 1 {
		s := os.Args[1]

		if _, err := strconv.ParseInt(s, 10, 64); err == nil {
			log.Printf("Using port %s", s)
			a.bindport = s
		}
	}

	// Create a string that will be used to make a connection later
	// Note Password has been left out, which is best to avoid issues when using null password
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	log.Println("Connecting to PostgreSQL")
	log.Println(psqlInfo)
	db, err := sql.Open("pgx", psqlInfo)
	a.db = db
	//db, err = sql.Open("sqlite3", "db.sqlite3")
	if err != nil {
		log.Println("Invalid DB arguments, or github.com/lib/pq not installed")
		log.Fatal(err)
	}

	// test connection
	err = a.db.Ping()
	if err != nil {
		log.Fatal("Connection to specified database failed: ", err)
	}

	log.Println("Database connected successfully")

	//check data import status
	_, err = os.Stat("./imported")
	if os.IsNotExist(err) {
		log.Println("--- Importing demo data")
		a.importData()
	}

	//set some defaults for the authentication to also support HTTP and HTTPS
	a.setupAuth()

	a.Router = mux.NewRouter()
	a.initializeRoutes()
}

func (a *App) initializeRoutes() {
	// setup static content route - strip ./assets/assets/[resource]
	// to keep /assets/[resource] as a route
	staticFileDirectory := http.Dir("./statics/")
	staticFileHandler := http.StripPrefix("/statics/", http.FileServer(staticFileDirectory))
	a.Router.PathPrefix("/statics/").Handler(staticFileHandler).Methods("GET")

	a.Router.HandleFunc("/", a.indexHandler).Methods("GET")
	a.Router.HandleFunc("/login", a.loginHandler).Methods("POST", "GET")
	a.Router.HandleFunc("/logout", a.logoutHandler).Methods("GET")
	a.Router.HandleFunc("/register", a.registerHandler).Methods("POST", "GET")
	a.Router.HandleFunc("/list", a.listHandler).Methods("GET")
	a.Router.HandleFunc("/list/{srt:[0-9]+}", a.listHandler).Methods("GET")
	a.Router.HandleFunc("/create", a.createHandler).Methods("POST", "GET")
	a.Router.HandleFunc("/update", a.updateHandler).Methods("POST", "GET")
	a.Router.HandleFunc("/delete", a.deleteHandler).Methods("POST", "GET")

	log.Println("Routes established")
	
}

func (a *App) Run(addr string) {
	if addr != "" {
		a.bindport = addr
	}

	// get the local IP that has Internet connectivity
	ip := GetOutboundIP()

	log.Printf("Starting HTTP service on http://%s:%s", ip, a.bindport)
	// setup HTTP on gorilla mux for a gracefull shutdown
	srv := &http.Server{
		//Addr: "0.0.0.0:" + a.bindport,
		Addr: ip + ":" + a.bindport,

		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      a.Router,
	}

	// HTTP listener is in a goroutine as its blocking
	go func() {
		if err = srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	// setup a ctrl-c trap to ensure a graceful shutdown
	// this would also allow shutting down other pipes/connections. eg DB
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	log.Println("shutting HTTP service down")
	srv.Shutdown(ctx)
	log.Println("closing database connections")
	a.db.Close()
	log.Println("shutting down")
	os.Exit(0)
}
