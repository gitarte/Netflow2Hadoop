package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
)

type config struct {
	host      string //	listen address
	port      string //	listen port
	chunkSize int    //	the size of accumulator
	dest      string //	destination directory
}

var c config

func main() {
	defer RecoverAnyPanic("main")

	//	Reading commandline arguments
	c.host = os.Args[1]                    //	listen address
	c.port = os.Args[2]                    //	listen port
	chunk, err := strconv.Atoi(os.Args[3]) //	the size of accumulator
	if err != nil {
		ExitOnError("main", err)
	}
	c.chunkSize = chunk
	c.dest = os.Args[4] //	destination directory

	//	Creating channel to pass netflow data betwean goroutines
	channel := make(chan []byte)

	//	Bringing to life infinite goroutine that accumulates netlows.
	//	It collects about 128MB of data and afterwards it sends this chunk
	//	into another goroutine that saves it into file
	go accumulate(channel)

	//	Creating UDP socket that will infinitely accept netlow traffic
	//	Each incomming message is handled in separate goroutine
	ServerAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%s", c.host, c.port))
	if err != nil {
		ExitOnError("ResolveUDPAddr", err)
	}
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	if err != nil {
		ExitOnError("ServerConn", err)
	}
	defer ServerConn.Close()
	fmt.Printf("Socket is listening on %s:%s\n", c.host, c.port)
	for {
		//	waiting for single transmission from netflow collector
		datagram := make([]byte, 4096)
		_, _, err := ServerConn.ReadFromUDP(datagram)
		if err != nil {
			LogOnError("ServerConn", err)
			continue
		}
		//	Sending data for for further processing
		//	which takes place in "accumulate" goroutine
		channel <- datagram
	}
}

func accumulate(channel chan []byte) {
	defer RecoverAnyPanic("accumulate")

	fileCount := 0               // integer counter of output files
	maxInt := int(^uint(0) >> 1) //	compute maximum value of type int

	for {
		//	prepare container for current data chunk
		chunk := make([][]byte, c.chunkSize)

		//	accumulate data
		i := 0
		for i < c.chunkSize {
			if i%100 == 0 {
				fmt.Printf("Accumulated %d flows in file %d\n", i, fileCount)
			}
			chunk[i] = <-channel // read data from UDP socket
			i++
		}

		//	store given data into file
		if fileCount < maxInt {
			fileCount++
		} else {
			fileCount = 0
		}
		go save(chunk, fileCount)
	}
}

func save(chunk [][]byte, fileCount int) {
	defer RecoverAnyPanic("save")

	//	create new file
	f, err := os.Create(fmt.Sprintf("%s/%s.flow", c.dest, strconv.Itoa(fileCount)))
	if err != nil {
		panic(err)
	}
	defer f.Close()

	//	feed file with JSON array of decoded flows
	f.WriteString("[")
	for _, datagram := range chunk {
		flow := decodeAsNetFlowV5(datagram)
		jSon, err := json.Marshal(&flow)
		if err != nil {
			LogOnError("ServerConn", err)
			continue
		}
		f.WriteString(fmt.Sprintf("%s,", string(jSon)))
	}
	f.WriteString("{}]") //	dirty trick that makes JSON always valid
}

func decodeAsNetFlowV5(buf []byte) NetFlowV5 {
	defer RecoverAnyPanic("decodeAsNetFlowV5")

	//	create new container for data
	var flow NetFlowV5
	//	Decode NetFlow v5 header
	flow.Header = DecodeHeader(buf)
	//	Decode NetFlow v5 records
	for c := uint16(0); c < flow.Header.Count; c++ {
		flow.Records = append(flow.Records, DecodeRecord(buf, c))
	}
	return flow
}
