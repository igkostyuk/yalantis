package main

import (
	"context"
	_ "expvar"
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/igkostyuk/yalantis/app/handlers"
	"github.com/igkostyuk/yalantis/internal/database"
	"github.com/pkg/errors"
)

// build is the git version of this program. It is set using build flags in the makefile.
var build = "develop"

func main() {
	log := log.New(os.Stdout, "COUNTER : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile|log.Ltime|log.LUTC)

	if err := run(log); err != nil {
		log.Println("main: error:", err)
		os.Exit(1)
	}
}

func run(log *log.Logger) error {

	// =========================================================================
	// Configuration
	apiHost := flag.String("api_host", "localhost:3000", "The ip:port for the api endpoint.")
	debugHost := flag.String("debug_host", "localhost:4000", "The ip:port for the debug endpoint.")
	readTimeout := flag.Duration("read_timeout", 5*time.Second, "The maximum duration for reading request.")
	writeTimeout := flag.Duration("write_timeout", 5*time.Second, "The maximum duration before timing out writes of the response.")
	shutdownTimeout := flag.Duration("sutdown_timeout", 5*time.Second, "The maximum duration for stop server gracefully.")

	dbHost := flag.String("db_host", "localhost:6379", "The ip:port for the api endpoint.")
	dbName := flag.String("db_name", "", "The name for database")
	dbUser := flag.String("db_user", "", "The user name for database")
	dbPassword := flag.String("db_password", "", "The password for database")

	flag.Parse()
	// =========================================================================
	// App Starting

	// =========================================================================
	// Start Database

	log.Println("main: Initializing database support")

	db, err := database.Open(database.Config{
		Host:     *dbHost,
		Name:     *dbName,
		User:     *dbUser,
		Password: *dbPassword,
	})

	if err != nil {
		return errors.Wrap(err, "connecting to db")
	}
	defer func() {
		log.Printf("main: Database Stopping : %s", *dbHost)
		db.Client.Close()
	}()

	// =========================================================================
	// Start Debug Service
	//
	// /debug/pprof - Added to the default mux by importing the net/http/pprof package.
	// /debug/vars - Added to the default mux by importing the expvar package.
	//
	// Not concerned with shutting this down when the application is shutdown.

	log.Println("main: Initializing debugging support")

	go func() {
		log.Printf("main: Debug Listening %s", *debugHost)
		if err := http.ListenAndServe(*debugHost, http.DefaultServeMux); err != nil {
			log.Printf("main: Debug Listener closed : %v", err)
		}
	}()

	// =========================================================================
	// Start API Service

	log.Println("main: Initializing API support")

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	api := http.Server{
		Addr:         *apiHost,
		Handler:      handlers.API(build, shutdown, log, db),
		ReadTimeout:  *readTimeout,
		WriteTimeout: *writeTimeout,
	}

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// Start the service listening for requests.
	go func() {
		log.Printf("main: API listening on %s", api.Addr)
		serverErrors <- api.ListenAndServe()
	}()

	// =========================================================================
	// Shutdown

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		return errors.Wrap(err, "server error")

	case sig := <-shutdown:
		log.Printf("main: %v : Start shutdown", sig)

		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(context.Background(), *shutdownTimeout)
		defer cancel()

		// Asking listener to shutdown and shed load.
		if err := api.Shutdown(ctx); err != nil {
			api.Close()
			return errors.Wrap(err, "could not stop server gracefully")
		}
	}

	return nil
}
