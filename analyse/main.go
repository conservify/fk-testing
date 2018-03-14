package main

import (
	"context"
	"flag"
	"log"
	"time"

	"encoding/json"
	_ "strings"

	_ "github.com/lib/pq"
	_ "github.com/paulmach/go.geo"

	"github.com/conservify/sqlxcache"

	"github.com/fieldkit/cloud/server/data"
)

type options struct {
	PostgresUrl string
}

type Record struct {
	ID        int64
	Timestamp time.Time
	Location  data.Location
	Data      []byte
}

func processAll(ctx context.Context, o *options, db *sqlxcache.DB) error {
	log.Printf("Querying...")

	updated := 0
	pageSize := 200
	page := 0
	for {
		batch := []*Record{}
		sql := `SELECT d.id, d.timestamp, ST_AsBinary(d.location) AS location, d.data FROM fieldkit.record AS d ORDER BY source_id, timestamp LIMIT $1 OFFSET $2`
		if err := db.SelectContext(ctx, &batch, sql, pageSize, page*pageSize); err != nil {
			log.Fatalf("Error: %v", err)
			panic(err)
		}

		if len(batch) == 0 {
			log.Printf("Done!")
			break
		}

		log.Printf("Processing %d records", len(batch))

		for _, record := range batch {
			data := make(map[string]interface{})
			err := json.Unmarshal(record.Data, &data)
			if err != nil {
				panic(err)
			}

			log.Printf("%v", data)
		}

		page += 1
	}

	log.Printf("Fixed %d rows", updated)

	return nil
}

func main() {
	o := options{}

	flag.StringVar(&o.PostgresUrl, "postgres-url", "postgres://fieldkit:password@127.0.0.1/fieldkit?sslmode=disable", "url to the postgres server")

	flag.Parse()

	db, err := sqlxcache.Open("postgres", o.PostgresUrl)
	if err != nil {
		panic(err)
	}

	ctx := context.TODO()

	err = processAll(ctx, &o, db)
	if err != nil {
		panic(err)
	}
}
