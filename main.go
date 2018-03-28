package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
)

func main() {
	//	reading commandline arguments
	host := os.Args[1] //	listen address
	port := os.Args[2] //	listen port
	//dstf := os.Args[3] //	directory for result files

	chanOfFlows := make(chan NetFlowV5)
	chanOfSlices := make(chan []NetFlowV5)

	//	creating UDP socket that will accept netlow traffic
	ServerAddr, _ := net.ResolveUDPAddr("udp", host+":"+port)
	ServerConn, _ := net.ListenUDP("udp", ServerAddr)
	defer ServerConn.Close()
	fmt.Printf("Socket listening on %s:%s\n", host, port)

	//	This goroutine accumulates netlows till it collects 500000 of them.
	//	Afterwards it sends accumulated data into another goroutine that saves it into file
	go func(channel chan NetFlowV5) {
		for {
			var flowsArray [500000]NetFlowV5

			i := 0
			for i < 500000 {
				flowsArray[i] = <-channel
				i++
			}

			chanOfSlices <- flowsArray[:]
		}
	}(chanOfFlows)

	//	This goroutine stores given data into file
	go func(channel chan []NetFlowV5) {
		fileCount := 0
		for {
			//	create new file
			fileCount++
			flowsArray := <-channel
			f, _ := os.Create("./" + strconv.Itoa(fileCount) + "flow")
			defer f.Close()

			//	feed file with JSON's
			for _, flow := range flowsArray {
				//	struct to json
				b, _ := json.Marshal(&flow)
				f.WriteString(string(b) + "\n")
			}
		}
	}(chanOfSlices)

	for {
		//	waiting for single transmission from netflow collector
		buf := make([]byte, 4096)
		n, _, _ := ServerConn.ReadFromUDP(buf)

		//	handling recaived data in separate goroutine
		go func(buf []byte, n int, channel chan NetFlowV5) {
			//	create new container for data
			var flow NetFlowV5
			//	Decode NetFlow v5 header
			flow.Header = DecodeHeader(buf)
			//	Decode NetFlow v5 records
			for c := uint16(0); c < flow.Header.Count; c++ {
				flow.Records = append(flow.Records, DecodeRecord(buf, c))
			}

			//	sending decoded netflow transmission into accumulating goroutine
			channel <- flow
		}(buf, n, chanOfFlows)
	}
}
