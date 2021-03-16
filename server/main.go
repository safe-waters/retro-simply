package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
)

//go:embed dist/*
var static embed.FS

func main() {
	p := os.Getenv("PORT")
	if p == "" {
		log.Fatal("PORT environment variable not set")
	}

	d, err := fs.Sub(static, "dist")
	if err != nil {
		log.Fatal(err)
	}

	b, err := fs.ReadFile(d, "retrospective.html")
	if err != nil {
		log.Fatal(err)
	}

	page := string(b)

	http.HandleFunc("/retrospective", servePage(page))
	http.Handle("/", http.FileServer(http.FS(d)))

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", p), nil))
}

func servePage(p string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, p)
	}
}
