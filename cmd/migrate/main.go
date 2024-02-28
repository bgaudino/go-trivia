package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"trivia/db"

	"github.com/bgaudino/godino"
	"github.com/jackc/pgx/v5"
)

func main() {
	fail := func(msg string, err error) {
		fmt.Fprintln(os.Stderr, err.Error())
		log.Fatal(msg)
	}
	pool, err := db.GetPool()
	if err != nil {
		fail("Could not connect to database", err)
	}
	ctx := context.Background()
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		fail("Error beginning transaction", err)
	}
	_, err = tx.Exec(
		ctx,
		`
			CREATE TABLE IF NOT EXISTS migrations (
				name CHAR(4) UNIQUE,
				applied TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`,
	)
	if err != nil {
		tx.Rollback(ctx)
		fail("Could not create migrations table", err)
	}
	appliedMigrations := godino.NewSet[string]()
	rows, err := tx.Query(ctx, "SELECT name FROM migrations")
	if err != nil {
		tx.Rollback(ctx)
		fail("Error querying migrations table", err)
	}
	for rows.Next() {
		var migration string
		err := rows.Scan(&migration)
		if err != nil {
			tx.Rollback(ctx)
			fail("Something went wrong", err)
		}
		appliedMigrations.Add(migration)
	}
	files, err := os.ReadDir("migrations")
	if err != nil {
		tx.Rollback(ctx)
		fail("Could not find migration files", err)
	}
	sort.Slice(files, func(i int, j int) bool { return files[i].Name() < files[j].Name() })
	files = godino.Filter(files, func(f os.DirEntry) bool { return !appliedMigrations.Has(f.Name()[:4]) })
	if len(files) == 0 {
		fmt.Println("No migrations to apply")
		return
	}
	for _, f := range files {
		if !appliedMigrations.Has(f.Name()[:4]) {
			c, ioErr := os.ReadFile(fmt.Sprintf("migrations/%v", f.Name()))
			if ioErr != nil {
				tx.Rollback(ctx)
				fail("Could not read migration file", ioErr)
			}
			sql := string(c)
			fmt.Printf("Applying %v...", f.Name())
			_, err := tx.Exec(ctx, sql)
			if err != nil {
				tx.Rollback(ctx)
				fail("Could not apply migration", err)
			}
			_, err = tx.Exec(ctx, "INSERT INTO migrations (name) VALUES ($1)", f.Name()[:4])
			if err != nil {
				tx.Rollback(ctx)
				fail("Error applying migration", err)
			}
			fmt.Println("OK")
		}
	}
	tx.Commit(ctx)
}
