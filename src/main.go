package main

import (
	//"fmt"
	"flag"
	"os/signal"
	"syscall"
	"os"
	"gopkg.in/redis.v3"
	"log"
	"sync"
	"time"
)

var (
	shard = flag.String("shard", "queue", "Name of the shard")
	redisHost = flag.String("host", "localhost", "Host for Redis")
	redisPort = flag.String("port", "3333", "Port for Redis")
	redisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		Password: "",
		DB: 0,
	})
	waitGroup = sync.WaitGroup{}
)

func RunJob(job string) {
	defer waitGroup.Done()

	time.Sleep(time.Second*10) // simulate work
	log.Printf("%s\n", job)

}

func HandleMessages(done chan bool, pubsub *redis.PubSub) {
	//defer waitGroup.Done()
	for {

		select {
		case <-done:
			log.Println("Stopping listener")
			return;
		default:
		}

		message, err := pubsub.ReceiveMessage()

		if err != nil {
			log.Fatalf("Error when receiving message: %v\n", err)
		}
		waitGroup.Add(1)
		go RunJob(message.Payload)
	}
}

func main() {
	log.Println("Queue Worker")
	log.Println("Written by Daniel Fekete <daniel.fekete@voov.hu>")
	flag.Parse()
	log.Printf("Running shard %s\n", *shard)

	_, err := redisClient.Ping().Result()
	if err != nil {
		log.Fatalf("Error when pinging Redis: %v\n", err)
	}

	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	pubsub, err := redisClient.Subscribe("billingo-1")
	if err != nil {
		log.Fatalf("Error when subscribing to shard %s: %v\n", shard,err)
	}
	done := make(chan bool)
	go HandleMessages(done, pubsub)
	<-signalCh
	close(done)

	// wait until every operation is finished
	waitGroup.Wait()

	pubsub.Close()
	redisClient.Close()

	log.Println("Queue Worker terminated :-(")
}
