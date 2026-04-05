package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/stockyard-dev/stockyard-estimate/internal/server"
	"github.com/stockyard-dev/stockyard-estimate/internal/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9802"
	}
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./estimate-data"
	}

	db, err := store.Open(dataDir)
	if err != nil {
		log.Fatalf("estimate: %v", err)
	}
	defer db.Close()

	srv := server.New(db, server.DefaultLimits())

	fmt.Printf("\n  Estimate — Self-hosted estimates and quotes with line items\n  Dashboard:  http://localhost:%s/ui\n  API:        http://localhost:%s/api\n  Questions? hello@stockyard.dev — I read every message\n\n", port, port)
	log.Printf("estimate: listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, srv))
}
