package pipelines

// ConvertSliceToClosedChannel creates a channel with capacity the same as the length of the passed in slice and adds
// the data from the slice into the channel in the same order as the slice. The channel is closed as no additional data
// should be added to it. This provides an easy way to essential queue data in a slice into a worker function.
func ConvertSliceToClosedChannel[T any](s []T) chan T {
	newChan := make(chan T, len(s))
	for _, v := range s {
		newChan <- v
	}
	close(newChan)
	return newChan
}
