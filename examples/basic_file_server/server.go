package main

import (
	"flag"
	"fmt"
	"github.com/ngmoco/falcore"
	"github.com/ngmoco/falcore/filter"
	"net/http"
)

// Command line options
var (
	port = flag.Int("port", 8000, "the port to listen on")
	path = flag.String("base", "./test", "the path to serve files from")
)

func main() {
	// parse command line options
	flag.Parse()

	// setup pipeline
	pipeline := falcore.NewPipeline()

	// set logger
	filter.SetLogger(falcore.GetLogger())

	// Serve index.html for root requests
	pipeline.Upstream.PushBack(filter.NewRequestFilter(func(req *filter.Request) *http.Response {
		if req.HttpRequest.URL.Path == "/" {
			req.HttpRequest.URL.Path = "/index.html"
		}
		return nil
	}))
	// Serve files
	pipeline.Upstream.PushBack(&filter.FileFilter{
		BasePath: *path,
	})

	// downstream
	pipeline.Downstream.PushBack(filter.NewCompressionFilter(nil))

	// setup server
	server := falcore.NewServer(*port, pipeline)

	// start the server
	// this is normally blocking forever unless you send lifecycle commands 
	if err := server.ListenAndServe(); err != nil {
		fmt.Println("Could not start server:", err)
	}
}
