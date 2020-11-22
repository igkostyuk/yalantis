// Package handlers contains the full set of handler functions and routes
// supported by the web api.
package handlers

import (
	"log"
	"net/http"
	"os"

	"github.com/igkostyuk/yalantis/app/data/count"
	"github.com/igkostyuk/yalantis/app/mid"
	"github.com/igkostyuk/yalantis/internal/database"
	"github.com/igkostyuk/yalantis/internal/web"
)

// API constructs an http.Handler with all application routes defined.
func API(build string, shutdown chan os.Signal, log *log.Logger, db database.DB) http.Handler {

	// Construct the web.App which holds all routes as well as common Middleware.
	app := web.NewApp(shutdown, mid.Logger(log), mid.Errors(log), mid.Metrics(), mid.Panics(log))

	//Register count endpoints.
	cg := countGroup{
		count: count.New(log, db),
	}
	app.Handle(http.MethodGet, "/", cg.get)

	return app
}
