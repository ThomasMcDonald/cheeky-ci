package main

import (
	"fmt"
	"net/http"
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

		_, err := fmt.Fprintln(w, "<p>test test test</p>")
		if err != nil {
			panic(err)
		}
	})
	fmt.Println("Server running at 8080")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
