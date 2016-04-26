package main

import (
	"fmt"
	"flag"
)

var (
	shard = flag.String("shard", "queue", "Name of the shard")
	redisHost = flag.String("host", "localhost", "Host for Redis")
	redisPort = flag.String("port", "3333", "Port for Redis")

)

func main() {
	fmt.Println("Queue Worker")
	fmt.Println("Written by Daniel Fekete <daniel.fekete@voov.hu>")
	flag.Parse()
}
