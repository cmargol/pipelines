package main

import (
	"context"
	"fmt"
	"github.com/cmargol/pipelines"
)

/*
 * This example highlights how a service can have a pipeline added to it; it is recommended to split up a service into
 * lower interfaces like this in case one section needs to be mocked while testing.
 */

const ServiceName = "ExampleService"

type ServiceDataType struct {
	// Image some data here
}

type TaggedData struct {
	data *ServiceDataType
	tags []string
}

type FetcherDecoder interface {
	Fetch(context.Context) ([]byte, error)
	Decode([]byte) (*ServiceDataType, error)
}

type Tagger interface {
	Tag(*ServiceDataType) (*TaggedData, error)
}

type EncoderPublisher interface {
	Encode(*TaggedData) ([]byte, error)
	Publish([]byte) error
}

type Service struct {
	fetchDecoder    FetcherDecoder
	tagger          Tagger
	encodePublisher EncoderPublisher
}

func NewService(fd FetcherDecoder, t Tagger, ep EncoderPublisher) Service {
	return Service{
		fetchDecoder:    fd,
		tagger:          t,
		encodePublisher: ep,
	}
}

// TagPipeline a pipeline that is responsible for consuming data, tagging it, and then publishing it.
func (s Service) TagPipeline(ctx context.Context) error {
	if s.fetchDecoder == nil {
		return fmt.Errorf("programming error nil interface, fetch decoder")
	}
	if s.tagger == nil {
		return fmt.Errorf("programming error nil interface, tagger")
	}
	if s.encodePublisher == nil {
		return fmt.Errorf("programming error nil interface, encode publisher")
	}
	shutdownErr := make(chan error, 1)

	queue, queueErrC := pipelines.Queue(ctx, s.fetchDecoder.Fetch, 1, 1)
	decodeC, decodeErrC := pipelines.WorkerPool(queue, s.fetchDecoder.Decode, 1, 1)
	taggedC, tagErrC := pipelines.WorkerPool(decodeC, s.tagger.Tag, 4, 2)
	encodedC, encodeErrC := pipelines.WorkerPool(taggedC, s.encodePublisher.Encode, 1, 1)
	dequeueC := pipelines.Dequeue(encodedC, s.encodePublisher.Publish, 1, 1)

	pipelineErrC := pipelines.Merge(queueErrC, decodeErrC, tagErrC, encodeErrC, dequeueC)

	go func() {
		for err := range pipelineErrC {
			switch e := err.(type) {
			case pipelines.ErrFatal:
				// shutdown pipeline if fatal
				shutdownErr <- e
			case pipelines.ErrPipeline:
				// perhaps just log and increment an error metric
			default:
				// shutdown pipeline if we are unsure
				shutdownErr <- e
			}
		}
	}()

	return <-shutdownErr
}
