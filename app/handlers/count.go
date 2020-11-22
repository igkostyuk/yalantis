package handlers

import (
	"context"
	"net/http"

	"github.com/igkostyuk/yalantis/app/data/count"
	"github.com/igkostyuk/yalantis/internal/web"
	"github.com/pkg/errors"
)

// // c contains the global counters.
// var globalCounter = expvar.NewInt("/ counts")

type countGroup struct {
	count count.Counter
}

func (cg countGroup) get(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if !ok {
		return web.NewShutdownError("web value missing from context")
	}

	counts, err := cg.count.Get(ctx, v.TraceID)
	if err != nil {
		return errors.Wrap(err, "unable to get count")
	}

	return web.Respond(ctx, w, counts, http.StatusOK)
}
