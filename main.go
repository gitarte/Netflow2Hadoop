package main

import (
	"fmt"
	"net"
)

func main() {
	ServerAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:9995")
	ServerConn, _ := net.ListenUDP("udp", ServerAddr)
	defer ServerConn.Close()

	for {
		buf := make([]byte, 4096)
		n, _, _ := ServerConn.ReadFromUDP(buf)
		if n < 24 {
			//	panic("malformed record")
			continue
		}
		go func(buf []byte, n int) {
			var flow NetFlowV5

			//	Decode NetFlow v5 header
			flow.Header = DecodeHeader(buf)

			//	Decode NetFlow v5 records
			for c := uint16(0); c < flow.Header.Count; c++ {
				// store decoded reckord
				flow.Records = append(flow.Records, DecodeRecord(buf, c))
			}

			fmt.Printf("%+v", flow)
		}(buf, n)
	}
}
