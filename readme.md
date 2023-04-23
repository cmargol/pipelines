# Pipelines
![Coverage](https://img.shields.io/badge/Coverage-93.3%25-brightgreen)

The purpose of pipelines module is to provide a library based solely off of the standard Golang library for common
patterns in data pipelines as well as some convenient function wrappers. 

## License
MIT License

## Usage
### Importing
`import "github.com/cmargol/pipelines"`

### When to use
* Need to set up concurrent workers to process data
* Part of your code could be defined by a flow chart or state diagram
* There is a bottleneck that is CPU blocking 

### When not to use
* Code does not require any concurrency / not CPU blocking

### Concurrency

### Error Handling Wrappers
This package comes with function wrappers that are capable of wrapping errors that occur in the Pipeline errors
that give functionality to give more context around what pipeline and what stage of the pipeline the error occurred.


### Metric Handling Wrappers
This package comes with several functions that can be used to wrap and provide metrics:
* `MetricWrapperQueue` : Wraps a queue function
* `MetricWrapperWorker` : Wraps a worker function 
* `MetricWrapperDequeue` : Wraps a dequeue function 

Each function takes in a `MetricsHandler` which is an interface with the following signatures:
* `RecordLastSuccessfulExecution(service string, stage string)`
* `RecordExecutionTime(t time.Duration, service string, stage string, status string)`
* `IncrementRecordCount(service string, stage string)`
* `IncrementErrorCount(service string, stage string)`

The wrappers automatically call the according functions when applied, however the actual implementation of 
the MetricsHandler is left up to choice to prevent locking down the implementation to one solution.

### Examples
* [Creating a Pipeline in a service](examples/pipeline-service.go)
* [Broadcasting](examples/broadcast.go)
* [Filtering Data](examples/filter-pipeline.go)
* [Error Wrapping](examples/filter-pipeline-with-helpers.go)
* [Queue using a channel](examples/queue-from-channel.go)
* [Queue using a slice](examples/queue-from-slice.go)

## Contributing
* Please create a branch with descriptive title and use the PR template included
* PR with any failing tests will not be reviewed and will automatically be rejected after 30 days