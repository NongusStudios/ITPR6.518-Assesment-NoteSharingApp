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
	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	wait time.Duration
)

type App struct {
	Router   *mux.Router
	db       *sql.DB
	bindport string
	//username string
	//role     string
}

func findBindPort() string {
	port := "8080"

	tempPort := os.Getenv("PORT")
	if tempPort != "" {
		port = tempPort
	}

	if len(os.Args) > 1 {
		s := os.Args[argBindport]

		if _, err := strconv.ParseInt(s, 10, 64); err == nil {
			port = s
		}
	}

	return port
}

func connectToPostgreSQL() (*sql.DB, error) {
	dbInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	log.Println(dbInfo)

	db, err := sql.Open("pgx", dbInfo)

	if err != nil {
		return nil, err
	}

	// Test DB connection
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func initRouter(a *App) *mux.Router {
	r := mux.NewRouter()

	staticFileDirectory := http.Dir("./statics/")
	staticFileHandler := http.StripPrefix("/statics/", http.FileServer(staticFileDirectory))
	r.PathPrefix("/statics/").Handler(staticFileHandler).Methods("GET")

	// Add handler functions
	r.HandleFunc("/", a.indexHandler).Methods("GET")
	r.HandleFunc("/login", a.loginHandler).Methods("POST", "GET")
	r.HandleFunc("/register", a.registerHandler).Methods("POST", "GET")
	r.HandleFunc("/logout", a.logoutHandler).Methods("GET")
	r.HandleFunc("/dashboard", a.dashboardHandler).Methods("GET")

	// Note handle
	r.HandleFunc("/search", a.searchHandler).Methods("POST")
	r.HandleFunc("/create", a.createNoteHandler).Methods("POST")
	r.HandleFunc("/edit", a.editNoteHandler).Methods("POST")
	r.HandleFunc("/delete", a.deleteNoteHandler).Methods("POST")
	r.HandleFunc("/editsettings", a.editSettingsHandler).Methods("POST")

	return r
}

func InitApp() (App, error) {
	a := App{}

	// Get the bindport
	a.bindport = findBindPort()
	log.Printf("Server using port %s\n", a.bindport)

	log.Println("Attempting to establish connection to PostgreSQL server")

	var err error
	a.db, err = connectToPostgreSQL()
	if err != nil {
		return App{}, err
	}

	// Check if tables imported
	_, err = os.Stat(fmt.Sprintf("./%s", dbFileLock))
	if os.IsNotExist(err) {
		a.importData()
	}

	log.Println("Successfully connected to PostgreSQL server")

	setupAuth()
	a.Router = initRouter(&a)

	return a, nil
}

func (a *App) Run() {
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
		if err := srv.ListenAndServe(); err != nil {
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
