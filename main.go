package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

// Config - global object that stores the configuration for entire application
var Config Configuration

func main() {
	defer RecoverAnyPanic("main")

	//	configure log file
	logFile := fmt.Sprintf("%s%s.log", os.Args[2], os.Args[0])
	f, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw)

	//	Read configuration file which is passed as first argument
	ReadConfig(&Config)

	//	Creating channel to pass decoded NetFlow data betwean goroutines
	JSONFlowChanel := make(chan string)

	if Config.Output.LocalFS.Enabled || Config.Output.HDFS.Enabled {
		//	Bringing to life infinite goroutine that accumulates netlows.
		//	It collects Config.ChunkSize of flows and afterwards it sends this
		//	into another goroutine that saves it into file
		go Accumulate(JSONFlowChanel)
	}

	if Config.Output.Kafka.Enabled {
		//	Bringing to life infinite goroutine that sends netlows into Kafka bus
		go SendingToKafka(JSONFlowChanel)
	}

	//	Creating UDP socket that will infinitely accept netlow traffic
	//	Each incoming message is handled in separate goroutine
	ServerAddr, err := net.ResolveUDPAddr("udp", Config.ListenParams)
	if err != nil {
		ExitOnError("ResolveUDPAddr", err)
	}
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	if err != nil {
		ExitOnError("ServerConn", err)
	}
	defer ServerConn.Close()
	log.Printf("Socket is listening on %s\n", Config.ListenParams)
	for {
		//	waiting for single transmission from netflow collector
		datagram := make([]byte, 4096)
		_, _, err := ServerConn.ReadFromUDP(datagram)
		if err != nil {
			LogOnError("ServerConn", err)
			continue
		}

		go func(datagram []byte) {
			flow := DecodeAsNetFlowV5(datagram)
			jSon, err := json.Marshal(&flow)
			if err != nil {
				LogOnError("main", err)
				return
			}
			JSONFlowChanel <- string(jSon)
		}(datagram)
	}
}

// Accumulate - infinite goroutine that accumulates NetFlows already decoded into JSON string
// It collects Config.ChunkSize of JSON strings ina slice and afterwards it sends the slice into
// another goroutine where it is saved into file
func Accumulate(JSONFlowChanel chan string) {
	defer RecoverAnyPanic("Accumulate")

	fileCount := 0               // integer counter of output files
	maxInt := int(^uint(0) >> 1) //	compute maximum value of type int

	for {
		//	prepare container for current data chunk
		chunk := make([]string, Config.Output.ChunkSize)

		//	accumulate data
		i := 0
		for i < Config.Output.ChunkSize {
			if i%100 == 0 {
				fmt.Printf("Accumulated %d flows in file %d\n", i, fileCount)
			}
			chunk[i] = <-JSONFlowChanel // read data from UDP socket
			i++
		}

		//	store given data into file
		if fileCount < maxInt {
			fileCount++
		} else {
			fileCount = 0
		}

		if Config.Output.LocalFS.Enabled {
			go SaveChunkToFile(chunk, fileCount)
		}

		if Config.Output.HDFS.Enabled {
			go SaveChunkToHDFS(chunk, fileCount)
		}
	}
}
