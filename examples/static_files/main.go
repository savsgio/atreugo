package main

import (
	"github.com/savsgio/atreugo/v9"
)

func main() {
	config := &atreugo.Config{
		Addr: "0.0.0.0:8000",
	}
	server := atreugo.New(config)

	// Register before middlewares
	server.UseBefore(beforeMiddleware)

	// Register after middlewares
	server.UseAfter(afterMiddleware)

	// Register a route with filters
	filters := atreugo.Filters{
		Before: []atreugo.Middleware{beforeFilter},
		After:  []atreugo.Middleware{afterFilter},
	}

	// Serve files with default configuration
	server.Static("/main", "./")

	// Serve just one file
	server.ServeFile("/readme", "README.md")

	// Serve just one file with filters
	server.ServeFileWithFilters("/license", "LICENSE", filters)

	// Creates a new group to serve static files
	static := server.NewGroupPath("/static")

	// Serves files with default configuration
	static.Static("/default", "./")

	// Serves files with default configuration and filters
	static.StaticWithFilters("/filters", "./", filters)

	// Serves files with your own custom configuration
	static.StaticCustom("/custom", &atreugo.StaticFS{
		Filters:            filters,
		Root:               "./",
		GenerateIndexPages: false,
		AcceptByteRange:    false,
		Compress:           true,
	})

	// Serve just one file
	static.ServeFile("/readme", "README.md")

	// Serve just one file with filters
	static.ServeFileWithFilters("/license", "LICENSE", filters)

	// Run
	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
