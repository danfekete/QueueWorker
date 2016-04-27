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
	//"time"
	"os/exec"
	"strings"
	"bytes"
	"github.com/hishboy/gocommons/lang"
	"time"
)

var (
	shard = flag.String("shard", "queue", "Name of the shard")
	redisHost = flag.String("host", "localhost:6379", "Host for Redis")
	workers = flag.Int("workers", 10, "Number of workers to use") // TODO: implement
	redisClient *redis.Client
	waitGroup = sync.WaitGroup{}
	jobQueue = lang.NewQueue() // FIFO list of jobs
)

/**
	Run a job
 */
func RunJob(job string) {
	defer waitGroup.Done()

	cmd := exec.Command(flag.Arg(0))
	cmd.Stdin = strings.NewReader(job)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("job returned: %q\n", out.String())
}


func Worker(id int, done chan bool) {
	for {
		select {
		case <-done:
			return;
		default:
		}

		jobString := jobQueue.Poll() // get job from the queue

		if jobString != nil {
			// there is a job to process
			time.Sleep(time.Second*5)
			log.Printf("Worked %d finished job %s\n", id, jobString)
			waitGroup.Done()
		}

	}
}

/**
	The main message handling function
 */
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
		//go RunJob(message.Payload)
		jobQueue.Push(message.Payload)
	}
}

func main() {
	log.Println("Queue Worker")
	log.Println("Written by Daniel Fekete <daniel.fekete@voov.hu>")
	flag.Parse()

	// We must have only one command line argument
	// besides the optional flags
	if flag.NArg() != 1 {
		log.Fatalln("Invalid number of arguments")
	}

	// connect to redis
	redisClient = redis.NewClient(&redis.Options{
		Addr: *redisHost,
		Password: "",
		DB: 0,
	})

	log.Printf("Running shard %s\n", *shard)

	_, err := redisClient.Ping().Result()
	if err != nil {
		log.Fatalf("Error when pinging Redis: %v\n", err)
	}

	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	// subscribe to redis channel
	pubsub, err := redisClient.Subscribe(*shard)
	if err != nil {
		log.Fatalf("Error when subscribing to shard %s: %v\n", shard,err)
	}

	done := make(chan bool) // channel to
	go HandleMessages(done, pubsub)

	for i:=1; i<=*workers; i++ {
		go Worker(i, done) // start 1 worker
	}


	<-signalCh // wait for the system signal
	close(done)

	// wait until every operation is finished
	waitGroup.Wait()

	pubsub.Close()
	redisClient.Close()

	log.Println("Queue Worker terminated :-(")
}
