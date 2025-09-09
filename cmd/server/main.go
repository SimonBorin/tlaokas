package main

import (
	"log"
	"net/http"

	"tlaokas/internal/cleanup"
	"tlaokas/internal/db"
	"tlaokas/internal/handler"
)

func main() {
	if err := db.Init(); err != nil {
		log.Fatal("DB init failed:", err)
	}
	go cleanup.Run()

	http.Handle("/secret", handler.WithCORS(http.HandlerFunc(handler.CreateSecret)))
	http.Handle("/secret/", handler.WithCORS(http.HandlerFunc(handler.GetSecret)))

	log.Println("Listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
