package count

import (
	"context"
	"log"
	"time"

	"github.com/igkostyuk/yalantis/internal/database"
	"github.com/pkg/errors"
)

type Counter struct {
	log *log.Logger
	db  database.DB
}

// New constructs a Count for api access.
func New(log *log.Logger, db database.DB) Counter {
	return Counter{
		log: log,
		db:  db,
	}
}

func (c Counter) Get(ctx context.Context, traceID string) (Count, error) {
	key := database.GenerateDBKey(time.Now())
	c.db.Client.SetNX(ctx, key, "0", 0)

	c.log.Printf("%s: %s: %s", traceID, "count.Get",
		database.Log(key),
	)

	count, err := c.db.Client.Incr(c.db.CTX, key).Result()
	if err != nil {
		return Count{}, errors.Wrap(err, "get count")
	}

	counts := Count{
		Counts: count,
	}

	return counts, nil
}
