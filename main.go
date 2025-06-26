package main

import (
	"embed"
	_ "embed"
	"io/fs"
	"log"
	"net/http"
	"tekuro/sentiment/sentiment"
	"text/template"
)

//go:embed templates/null_page.html
var nullPage []byte

//go:embed static/*
var static embed.FS

func main() {
	mux := http.NewServeMux()

	openaiClient, err := sentiment.NewOpenAi()
	if err != nil {
		log.Fatalf("Failed to start server, could not construct openai client: %v", err)
	}

	mux.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write(nullPage)
	})

	tmpl, err := template.ParseFiles("templates/sentiment.html")
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}
	mux.HandleFunc("GET /get/{ticker}", func(w http.ResponseWriter, r *http.Request) {
		data := struct{ Ticker string }{Ticker: r.PathValue("ticker")}
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("GET /sse/{ticker}", func(w http.ResponseWriter, r *http.Request) {
		ticker := r.PathValue("ticker")
		if sse, err := sentiment.NewSSEWriter(w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			openaiClient.Sentiment(r.Context(), ticker, sse)
		}
	})

	staticFiles, err := fs.Sub(static, "static")
	if err != nil {
		log.Fatal(err)
	}
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFiles))))

	log.Fatal(http.ListenAndServe(":5000", mux))
}
