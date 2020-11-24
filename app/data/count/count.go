package count

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/igkostyuk/yalantis/internal/database"
	"github.com/pkg/errors"
)

// c contains the global counters.
//var globalCounter = expvar.NewInt("/ counts")

const delayUpdatingDB = 10 * time.Second

//Counter for api access.
type Counter struct {
	log        *log.Logger
	db         database.DB
	count      int64
	updatingDB bool

	mu sync.RWMutex
}

// New constructs a Counter for api access.
func New(log *log.Logger, db database.DB) Counter {

	count := dbGet(log, db)

	return Counter{
		log:        log,
		db:         db,
		updatingDB: false,
		count:      count,
	}
}

// Get return count and update database.
func (c *Counter) Get(ctx context.Context, traceID string) (Count, error) {

	c.mu.Lock()
	defer c.mu.Unlock()
	c.count++

	//count := atomic.AddInt64(&c.count, 1)
	count := c.count

	c.log.Printf("%s: %s", traceID, "count.Get")

	counts := Count{
		Counts: count,
	}

	if !c.updatingDB {
		go c.UpdateDB(traceID)
		c.updatingDB = true
	}

	return counts, nil
}

// UpdateDB updated database
func (c *Counter) UpdateDB(traceID string) error {
	<-time.After(delayUpdatingDB)

	c.mu.Lock()
	defer c.mu.Unlock()

	key := database.GenerateDBKey(time.Now())
	c.db.Client.SetNX(c.db.CTX, key, "0", 0)

	c.log.Printf("%s: %s: %s", traceID, "count.UpdateDB",
		database.Log(key),
	)
	count := c.count

	err := c.db.Client.Set(c.db.CTX, key, count, 0).Err()
	if err != nil {
		return errors.Wrap(err, "update count")
	}

	c.updatingDB = false

	return nil
}

func dbGet(log *log.Logger, db database.DB) int64 {
	key := database.GenerateDBKey(time.Now())
	isNew, err := db.Client.SetNX(db.CTX, key, "0", 0).Result()
	if err != nil {
		log.Printf("%s: %s: %s: %s",
			"count.init", "can't create new key", key, err)
	}
	if isNew {
		log.Printf("%s: %s: %s:",
			"count.init", "created new key", key)
		return 0
	}
	count, err := db.Client.Get(db.CTX, key).Int64()
	if err != nil {
		log.Printf("%s: %s: %s: %s",
			"count.init", "can't get key", key, err)
		return 0
	}
	log.Printf("%s: %s: %d",
		"count.init", "set global counter to", count)

	return count
}
