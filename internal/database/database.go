// Package database provides support for access the database.
package database

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type DB struct {
	Client *redis.Client
	CTX    context.Context
}

// Config is the required properties to use the database.
type Config struct {
	User       string
	Password   string
	Host       string
	Name       string
	DisableTLS bool
}

// Open knows how to open a database connection based on the configuration.
func Open(cfg Config) (DB, error) {
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Host,
		Password: cfg.Password, // no password set
		DB:       0,
	})

	db := DB{
		Client: rdb,
		CTX:    ctx,
	}
	_, err := rdb.Ping(ctx).Result()

	key := GenerateDBKey(time.Now())
	rdb.SetNX(ctx, key, "0", 0)
	return db, err
}

// StatusCheck returns nil if it can successfully talk to the database. It
// returns a non-nil error otherwise.
func StatusCheck(ctx context.Context, db *redis.Client) error {

	pong, err := db.Ping(ctx).Result()
	if err != nil && pong != "" {
		return err
	}
	return nil
}

// Log provides a pretty print version of the query and parameters.
func Log(key string) string {
	return key
}

func GenerateDBKey(t time.Time) string {
	return t.UTC().Format("January 2, 2006")
}
