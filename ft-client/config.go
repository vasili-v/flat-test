package main

import "flag"

var (
	server  string
	total   int
	clients int
	size    int
)

func init() {
	flag.StringVar(&server, "s", "127.0.0.1:5555", "address:port of server")
	flag.IntVar(&total, "n", 5, "number of requests to send")
	flag.IntVar(&clients, "c", 1, "number of parallel gRPC streams")
	flag.IntVar(&size, "a", 4, "number of attributes")

	flag.Parse()
}
