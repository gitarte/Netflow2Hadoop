package main

import (
	"fmt"
	"os"
	"strconv"
)

// SaveChunkToFile - goroutine that that performs saving data to file on local file system
// It is brought to life each time there is completed chunk of data in accumulator
func SaveChunkToFile(chunk []string, fileCount int) {
	defer RecoverAnyPanic("SaveChunkToFile")

	//	create new file
	f, err := os.Create(fmt.Sprintf("%s/%sflow.json", Config.Output.LocalFS.Path, strconv.Itoa(fileCount)))
	if err != nil {
		LogOnError("SaveChunkToFile", err)
		return
	}
	defer f.Close()

	//	feed file with JSON array of decoded flows
	f.WriteString("[")
	for _, jSon := range chunk {
		f.WriteString(fmt.Sprintf("%s,", jSon))
	}
	f.WriteString("{}]") //	dirty trick that makes JSON always valid
}
