package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
)

// C - global object that stores the configuration for entire application
var C Config

func main() {
	defer RecoverAnyPanic("main")

	//	Read configuration file which is passed as first argument
	ReadConfig(&C)
	fmt.Printf("%+v\n", C)

	//	Creating channel to pass netflow data betwean goroutines
	channel := make(chan []byte)

	//	Bringing to life infinite goroutine that accumulates netlows.
	//	It collects about 128MB of data and afterwards it sends this chunk
	//	into another goroutine that saves it into file
	go accumulate(channel)

	//	Creating UDP socket that will infinitely accept netlow traffic
	//	Each incomming message is handled in separate goroutine
	ServerAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", C.Host, C.Port))
	if err != nil {
		ExitOnError("ResolveUDPAddr", err)
	}
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	if err != nil {
		ExitOnError("ServerConn", err)
	}
	defer ServerConn.Close()
	fmt.Printf("Socket is listening on %s:%d\n", C.Host, C.Port)
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
		chunk := make([][]byte, C.ChunkSize)

		//	accumulate data
		i := 0
		for i < C.ChunkSize {
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
	f, err := os.Create(fmt.Sprintf("%s/%s.flow", C.Dest, strconv.Itoa(fileCount)))
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
		fmt.Println(string(jSon))
		f.WriteString(fmt.Sprintf("%s,", string(jSon)))
	}
	f.WriteString("{}]") //	dirty trick that makes JSON always valid
}

func decodeAsNetFlowV5(buf []byte) NetFlowV5 {
	defer RecoverAnyPanic("decodeAsNetFlowV5")

	//	create new container for data
	var flow NetFlowV5
	//	Decode NetFlowV5 header
	flow.Header = DecodeHeader(buf)
	//	Decode NetFlowV5 records
	flow.Records = make([]Record, flow.Header.Count)
	for i := uint16(0); i < flow.Header.Count; i++ {
		recBuf := buf[i*RecordLength+24 : i*RecordLength+24+RecordLength]
		flow.Records[i] = DecodeRecord(recBuf)
	}
	return flow
}
