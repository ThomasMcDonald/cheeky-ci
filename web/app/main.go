package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, err := fmt.Fprintln(w, `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>HTMX Demo</title>
<script src="https://unpkg.com/htmx.org"></script>
</head>
<body>
<h1>HTMX Demo</h1>
<div id="content">
<p>Click the button to fetch updated content!</p>
</div>
<button hx-get="/update" hx-target="#content" hx-swap="innerHTML">
Get Updated Content
</button>
</body>
</html>`)
		if err != nil {
			panic(err)
		}
	})

	http.HandleFunc("/update", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
		if err != nil {
			fmt.Fprintln(os.Stderr, "Unable to connect to database %v\n", err)
		}

		defer conn.Close(context.Background())

		name := "initial job"
		var id string

		err = conn.QueryRow(context.Background(), "INSERT INTO jobs (name) VALUES ($1) RETURNING id", name).Scan(&id)
		if err != nil {
			fmt.Fprintln(os.Stderr, "query error %v\n", err)
		}

		_, err = fmt.Fprintln(w, "<p>New row: %s</p>", id)
		if err != nil {
			panic(err)
		}
	})

	http.HandleFunc("/runner/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		_, err := fmt.Fprintln(w, "{}")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/runner/{id}/heartbeat", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		_, err := fmt.Fprintln(w, "{}")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})

	fmt.Println("Server running at 5173")

	err := http.ListenAndServe(":5173", nil)
	if err != nil {
		panic(err)
	}
}
