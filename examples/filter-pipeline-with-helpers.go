package main

import (
	"context"
	"fmt"
	"github.com/cmargol/pipelines"
	"log"
	"math/rand"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	generateRandomColors := func(ctx context.Context) (string, error) {
		choices := []string{"red", "green", "blue"}
		choice := rand.Int() % len(choices)
		return choices[choice], nil
	}
	filterColorRed := func(s string) (string, error) {
		if s == "red" {
			return "", nil
		}
		return s, nil
	}
	printColor := func(s string) error {
		fmt.Println(s)
		return nil
	}

	queueC, queueErrC := pipelines.Queue(ctx,
		pipelines.QueueFunctionErrWrapper(generateRandomColors, "colors", "generator"),
		1,
		1)
	filteredC, filteredErrC := pipelines.WorkerPoolWithZeroValueFilter(
		queueC,
		pipelines.WorkerFunctionErrWrapper(filterColorRed, "colors", "filter"),
		2,
		2,
	)
	dequeueErrC := pipelines.Dequeue(
		filteredC,
		pipelines.DequeueFunctionErrWrapper(printColor, "colors", "printer"),
		1,
		1)

	for err := range pipelines.Merge(queueErrC, filteredErrC, dequeueErrC) {
		log.Fatal(err)
	}
}
