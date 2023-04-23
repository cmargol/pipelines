package main

import (
	"context"
	"fmt"
	"github.com/cmargol/pipelines"
	"log"
	"math/rand"
)

type Transaction struct {
	Name           string
	Amount         float64
	Classification string
}

func (t Transaction) String() string {
	return fmt.Sprintf("Name: %s , Amount: %f", t.Name, t.Amount)
}

func main() {
	ctx := context.TODO()
	newTransaction := func(ctx context.Context) (Transaction, error) {
		names := []string{"KFC", "Bojangles", "Pizza Hut", "Golang Tutorial"}
		name := names[rand.Int()%len(names)]
		amount := float64(int(rand.Float64()*100.0)%5000) / 100.0
		return Transaction{Name: name, Amount: amount}, nil
	}

	// Create channels we will fan out to...
	txLoggerChan := make(chan Transaction)
	txClassifierChan := make(chan Transaction)

	// Queue data
	txQueueChan, txQueueErrChan := pipelines.Queue(ctx, newTransaction, 1, 1)
	// Broadcast and set the same data from the first channel to rest of the channels
	pipelines.Broadcast(txQueueChan, txLoggerChan, txClassifierChan)

	loggerErrChan := pipelines.Dequeue(txLoggerChan, func(transaction Transaction) error {
		fmt.Printf("%s\n", transaction)
		return nil
	}, 1, 1)

	classifiedTxChan, classificationErrChan := pipelines.WorkerPool(txClassifierChan, func(transaction Transaction) (Transaction, error) {
		newTx := transaction
		if transaction.Amount >= 50.00 {
			newTx.Classification = "high"
			return newTx, nil
		}
		newTx.Classification = "low"
		return newTx, nil
	}, 2, 2)

	voidDequeueErrC := pipelines.Dequeue(classifiedTxChan, func(transaction Transaction) error {
		return nil
	}, 1, 1)

	for err := range pipelines.Merge(txQueueErrChan, loggerErrChan, classificationErrChan, voidDequeueErrC) {
		log.Println("[ERROR] ", err)
	}
}
