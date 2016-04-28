# QueueWorker

Subscribes to `shard` channel on Redis and pushes every incoming jobs to a Queue. Prespecified number of workers take jobs out of the queue and send it to the command's stdin.

When job is finished it publishes the command's stdout to `pub` channel on Redis

### Usage

 ```
qworker [options] "command"
Options:
  -host string
        Host for Redis (default "localhost:6379")
  -pub string
        Name of the channel to publish the command output to (default "ws")
  -shard string
        Name of the shard (default "queue")
  -workers int
        Number of workers to use (default 10)
 ```
