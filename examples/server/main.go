package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/gorules/zen-go"
)

//go:embed rules
var rulesFS embed.FS

func readTestFile(key string) ([]byte, error) {
	data, err := rulesFS.ReadFile(path.Join("rules", key))
	if err != nil {
		return nil, err
	}

	return data, nil
}

func main() {
	// Create a new engine instance
	engine := zen.NewEngine(zen.EngineConfig{Loader: readTestFile})
	defer engine.Dispose()

	// Create a new server
	server := zen.NewServer(engine)

	// Get the port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start the server
	fmt.Printf("Starting server on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, server.Routes()); err != nil {
		log.Fatal(err)
	}
}
