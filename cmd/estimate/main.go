package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/stockyard-dev/stockyard-estimate/internal/server"
	"github.com/stockyard-dev/stockyard-estimate/internal/store"
	"github.com/stockyard-dev/stockyard/bus"
)

var version = "dev"

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("FATAL PANIC: %v\n%s", r, debug.Stack())
			os.Exit(1)
		}
	}()

	portFlag := flag.String("port", "", "HTTP port")
	dataFlag := flag.String("data", "", "Data directory for SQLite files")
	flag.Parse()

	port := *portFlag
	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = "9802"
	}

	dataDir := *dataFlag
	if dataDir == "" {
		dataDir = os.Getenv("DATA_DIR")
	}
	if dataDir == "" {
		dataDir = "./estimate-data"
	}

	db, err := store.Open(dataDir)
	if err != nil {
		log.Fatalf("estimate: %v", err)
	}
	defer db.Close()

	// Bus: one level up from the private data dir so every tool in a
	// bundle shares one _bus.db. Non-fatal.
	var b *bus.Bus
	if bb, berr := bus.Open(filepath.Dir(dataDir), "estimate"); berr != nil {
		log.Printf("estimate: bus disabled: %v", berr)
	} else {
		b = bb
		defer b.Close()
	}

	srv := server.New(db, server.DefaultLimits(dataDir), dataDir, b)

	fmt.Printf("\n  Estimate v%s — Self-hosted estimates and quotes with line items\n  Dashboard:  http://localhost:%s/ui\n  API:        http://localhost:%s/api\n  Data:       %s\n  Questions? hello@stockyard.dev — I read every message\n\n", version, port, port, dataDir)
	log.Printf("estimate: listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, srv))
}
