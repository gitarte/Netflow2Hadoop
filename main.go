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
	defer RecoverAnyPanic("main")

	//	Reading commandline arguments
	c.host = os.Args[1]                    //	listen address
	c.port = os.Args[2]                    //	listen port
	chunk, err := strconv.Atoi(os.Args[3]) //	the size of accumulator
	if err != nil {
		ExitOnError("main", err)
	}
	c.chunk = chunk
	c.dest = os.Args[4] //	destination directory

	//	Creating channel to pass netflow data betwean goroutines
	chanOfFlows := make(chan NetFlowV5)

	//	Bringing to life infinite goroutine that accumulates netlows.
	//	It collects about 128MB of data and afterwards it sends this chunk
	//	into another goroutine that saves it into file
	go accumulate(chanOfFlows)

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
		buf := make([]byte, 4096)
		n, _, err := ServerConn.ReadFromUDP(buf)
		if err != nil {
			LogOnError("ServerConn", err)
			continue
		}
		//	decoding recaived data in separate goroutine
		go decodeDatagrams(buf, n, chanOfFlows)
	}
}

func decodeDatagrams(buf []byte, n int, channel chan NetFlowV5) {
	defer RecoverAnyPanic("decodeDatagrams")

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
	defer RecoverAnyPanic("accumulate")

	fileCount := 0
	for {
		//	prepare container for current data chunk
		sliceOfFlows := make([]NetFlowV5, c.chunk)
		fileCount++

		//	accumulate data
		i := 0
		for i < c.chunk {
			if i%100 == 0 {
				fmt.Printf("Accumulated %d flows in file %d\n", i, fileCount)
			}

			sliceOfFlows = append(sliceOfFlows, <-channel)
			i++
		}

		//	This goroutine stores given data into file
		go save(sliceOfFlows, fileCount)
	}
}

func save(sliceOfFlows []NetFlowV5, fileCount int) {
	defer RecoverAnyPanic("save")

	//	create new file
	f, _ := os.Create(fmt.Sprintf("%s/%s.flow", c.dest, strconv.Itoa(fileCount)))
	defer f.Close()

	//	feed file with JSON array of decoded flows
	f.WriteString("[")
	for _, flow := range sliceOfFlows {
		b, _ := json.Marshal(&flow)
		f.WriteString(fmt.Sprintf("%s,", string(b)))
	}
	f.WriteString("{}]") //	dirty trick that makes JSON always valid
}
