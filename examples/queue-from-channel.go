package main

import (
	"fmt"
	"github.com/cmargol/pipelines"
	"log"
	"time"
)

func main() {
	queueC := make(chan string, 4)
	queueC <- "sleep"
	queueC <- "wait"
	queueC <- "wake-up"
	queueC <- "attack"
	close(queueC)
	// Sets up the slice of commands to be queued up
	// Creates a worker that pulls off the slice like it is a queue
	dequeueErrC := pipelines.Dequeue(queueC, func(s string) error {
		switch s {
		case "sleep":
			fmt.Println("You take a nap for 15 seconds")
			time.Sleep(15 * time.Second)
		case "wake-up":
			fmt.Println("You have woken up")
		case "attack":
			fmt.Println("You swing your sword towards the enemy")
		default:
			return fmt.Errorf("unrecognized command %s", s)
		}
		return nil
	}, 1, 1)
	for err := range dequeueErrC {
		log.Println(err)
	}
}
