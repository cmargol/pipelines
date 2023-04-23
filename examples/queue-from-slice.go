package main

import (
	"fmt"
	"github.com/cmargol/pipelines"
	"log"
	"time"
)

func main() {
	commands := []string{
		"sleep",
		"wait",
		"wake-up",
		"attack",
	}
	// Sets up the slice of commands to be queued up
	queueC := pipelines.ConvertSliceToClosedChannel(commands)
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
