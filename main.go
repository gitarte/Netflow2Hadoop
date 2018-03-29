package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
)

type config struct {
	host  string //	listen address
	port  string //	listen port
	chunk int    //	the size of accumulator
	dest  string //	destination directory
}

var c config

func main() {
	//	Reading commandline arguments
	c.host = os.Args[1]                   //	listen address
	c.port = os.Args[2]                   //	listen port
	c.chunk, _ = strconv.Atoi(os.Args[3]) //	the size of accumulator
	c.dest = os.Args[4]                   //	destination directory

	//	Creating channel to pass netflow data betwean goroutines
	chanOfFlows := make(chan NetFlowV5)

	//	Bringing to life infinite goroutine that accumulates netlows.
	//	It collects about 128MB of data and afterwards it sends this chunk
	//	into another goroutine that saves it into file
	go accumulate(chanOfFlows)

	//	Creating UDP socket that will infinitely accept netlow traffic
	//	Each incomming message is handled in separate goroutine
	ServerAddr, _ := net.ResolveUDPAddr("udp", c.host+":"+c.port)
	ServerConn, _ := net.ListenUDP("udp", ServerAddr)
	defer ServerConn.Close()
	fmt.Printf("Socket is listening on %s:%s\n", c.host, c.port)
	for {
		//	waiting for single transmission from netflow collector
		buf := make([]byte, 4096)
		n, _, _ := ServerConn.ReadFromUDP(buf)

		//	decoding recaived data in separate goroutine
		go decodeDatagrams(buf, n, chanOfFlows)
	}
}

func decodeDatagrams(buf []byte, n int, channel chan NetFlowV5) {
	//	create new container for data
	var flow NetFlowV5
	//	Decode NetFlow v5 header
	flow.Header = DecodeHeader(buf)
	//	Decode NetFlow v5 records
	for c := uint16(0); c < flow.Header.Count; c++ {
		flow.Records = append(flow.Records, DecodeRecord(buf, c))
	}

	//	sending decoded netflow into accumulating goroutine
	channel <- flow
}

func accumulate(channel chan NetFlowV5) {
	fileCount := 0
	for {
		//	prepare container for current data chunk
		sliceOfFlows := make([]NetFlowV5, c.chunk)
		fileCount++

		//	accumulate data
		i := 0
		for i < c.chunk {
			sliceOfFlows = append(sliceOfFlows, <-channel)
			i++
		}

		//	This goroutine stores given data into file
		go save(sliceOfFlows, fileCount)
	}
}

func save(sliceOfFlows []NetFlowV5, fileCount int) {
	//	create new file
	f, _ := os.Create(c.dest + "/" + strconv.Itoa(fileCount) + ".flow")
	defer f.Close()

	//	feed file with JSON array of decoded flows
	f.WriteString("[")
	for _, flow := range sliceOfFlows {
		b, _ := json.Marshal(&flow)
		f.WriteString(string(b) + ",")
	}
	f.WriteString("{}]") //	dirty trick that makes JSON always valid
}
