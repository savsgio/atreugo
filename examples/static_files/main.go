package main

import (
	"github.com/savsgio/atreugo/v10"
)

func main() {
	config := atreugo.Config{
		Addr: "0.0.0.0:8000",
	}
	server := atreugo.New(config)

	// Register before middlewares
	server.UseBefore(beforeGlobal)

	// Register after middlewares
	server.UseAfter(afterGlobal)

	// Middlewares collection
	middlewares := atreugo.Middlewares{
		Before: []atreugo.Middleware{beforeView},
		After:  []atreugo.Middleware{afterView},
	}

	// Serve files with default configuration
	server.Static("/main", "./")

	// Serve just one file
	server.ServeFile("/readme", "README.md")

	// Serve just one file with middlewares
	server.ServeFile("/license", "LICENSE").Middlewares(middlewares)

	// Creates a new group to serve static files
	static := server.NewGroupPath("/static")

	// Serves files with default configuration
	static.Static("/default", "./")

	// Serves files with default configuration and middlewares
	static.Static("/middlewares", "./").Middlewares(middlewares)

	// Serves files with your own custom configuration
	static.StaticCustom("/custom", &atreugo.StaticFS{
		Root:               "./",
		GenerateIndexPages: false,
		AcceptByteRange:    false,
		Compress:           true,
	}).SkipMiddlewares(beforeGlobal)

	// Serve just one file
	static.ServeFile("/readme", "README.md").UseBefore(beforeView)

	// Serve just one file with middlewares
	static.ServeFile("/license", "LICENSE").Middlewares(middlewares)

	// Run
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
