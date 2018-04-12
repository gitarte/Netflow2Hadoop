package main

// SaveChunkToHDFS - goroutine that that performs saving data to file on HDFS
// It is brought to life each time there is completed chunk of data in accumulator
func SaveChunkToHDFS(chunk []string, fileCount int) {
	defer RecoverAnyPanic("SaveChunkToHDFS")

	// This was fucked up
	// TODO: use https://github.com/colinmarc/hdfs
}
