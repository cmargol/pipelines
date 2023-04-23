// Package pipelines is meant to provide a set of generic functions to help provide error free consistent code
// for data pipelines. The following pipeline concepts that will be implemented within this package:
//
//   - Queue : Function that can be called multiple times to produce the next data to work with (Ex. getting next value
//     from slice)
//
// - WorkerPool : Stage in pipeline that will do some work on incoming data and return some data and an error
//
// - Dequeue : Termination function that is called for the last step in our pipeline ex (Ex. Storing to database, etc.)
//
// Outside of these general three functions will be variants of each and wrapper functions to help aid functionality
// such as logging, metrics, and error handling.
//
// Each step of the pipeline will return an error channel; It is expected that all of these errors will be
// merged together and handled it an error handling function
package pipelines
