package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/andrianbdn/sqembv"
)

const version = "0.1.1"

var options struct {
	db   string
	host string
	port uint
}

// printHeader print the welcome header.
func printHeader() {
	_, _ = fmt.Fprintf(os.Stderr, "sqembv, v%s\n", version)
}

func initConfig() {
	const defaultDB = "test.sqlite3"

	// Define flags
	flag.StringVar(&options.db, "db", defaultDB, "SQLite database file")
	flag.StringVar(&options.host, "bind", "localhost", "HTTP server host")
	flag.UintVar(&options.port, "listen", 8000, "HTTP server listen port")
	printLicense := false
	flag.BoolVar(&printLicense, "license", false, "Print program license and exit")

	// Parse flags
	flag.Parse()

	if printLicense {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", sqembv.License())
		os.Exit(0)
	}

	dbExists := fileExists(options.db)

	// Check if db is provided
	if options.db == "" || (options.db == defaultDB && !dbExists) {
		flag.PrintDefaults() // Print the usage of all defined flags
		os.Exit(1)
	}

	if !dbExists {
		_, _ = fmt.Fprintf(os.Stderr, "Error: db file %s does not exist\n", options.db)
		os.Exit(1)
	}
}

func fileExists(filePath string) bool {
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// startServer initialize and start the web server.
func startServer() {
	api, err := sqembv.NewAPI(options.db)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Error: cannot open db", err)
		os.Exit(1)
	}

	_, _ = fmt.Fprintf(os.Stderr, "serving view for %s at http://%s:%d/\n",
		options.db, options.host, options.port)

	http.Handle("/", api.Handler("/"))

	err = http.ListenAndServe(fmt.Sprintf("%s:%d", options.host, options.port), nil)

	_, _ = fmt.Fprintf(os.Stderr, "error: %s\n", err)
}

func main() {
	printHeader()
	initConfig()
	startServer()
}
