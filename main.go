package main

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

type netFlowV5Header struct {
	Version          uint16 `json:"Version"`          // 00-01	version				NetFlow export format version number
	Count            uint16 `json:"Count"`            // 02-03	count				Number of flows exported in this packet (1-30)
	SysUptime        int32  `json:"sysUptime"`        // 04-07	sys_uptime			Current time in milliseconds since the export device booted
	Timestamp        string `json:"Timestamp"`        // 08-11	unix_secs			Current count of seconds since 0000 UTC 1970 and 12-15	unix_nsecs	Residual nanoseconds since 0000 UTC 1970
	FlowSequence     int32  `json:"FlowSequence"`     // 16-19	flow_sequence		Sequence counter of total flows seen
	EngineType       uint8  `json:"EngineType"`       // 20		engine_type			Type of flow-switching engine
	EngineID         uint8  `json:"EngineID"`         // 21		engine_id			Slot number of the flow-switching engine
	SamplingInterval uint16 `json:"SamplingInterval"` // 22-23	sampling_interval	First two bits hold the sampling mode; remaining 14 bits hold value of sampling interval
}

type netFlowV5Record struct {
	SrcAddr string `json:"SrcAddr"` // 00-03	srcaddr	Source IP address
	DstAddr string `json:"DstAddr"` // 04-07	dstaddr	Destination IP address
	NextHop string `json:"NextHop"` // 08-11	nexthop	IP address of next hop router
	Input   uint16 `json:"Input"`   // 12-13	input	SNMP index of input interface
	Output  uint16 `json:"Output"`  // 14-15	output	SNMP index of output interface
	DPkts   uint32 `json:"DPkts"`   // 16-19	dPkts	Packets in the flow
	DOctets uint32 `json:"DOctets"` // 20-23	dOctets	Total number of Layer 3 bytes in the packets of the flow
	First   uint32 `json:"First"`   // 24-27	first	SysUptime at start of flow
	Last    uint32 `json:"Last"`    // 28-31	last	SysUptime at the time the last packet of the flow was received
	SrcPort uint16 `json:"SrcPort"` // 32-33	srcport	TCP/UDP source port number or equivalent
	DstPort uint16 `json:"DstPort"` // 34-35	dstport	TCP/UDP destination port number or equivalent
	// 36	pad1	Unused (zero) bytes
	TCPFlags uint8  `json:"TCPFlags"` // 37		tcp_flags	Cumulative OR of TCP flags
	Prot     uint8  `json:"Prot"`     // 38		prot		IP protocol type (for example, TCP = 6; UDP = 17)
	Tos      uint8  `json:"Tos"`      // 39		tos	IP 		type of service (ToS)
	SrcAs    uint16 `json:"SrcAs"`    // 40-41	src_as		Autonomous system number of the source, either origin or peer
	DstAs    uint16 `json:"DstAs"`    // 42-43	dst_as		Autonomous system number of the destination, either origin or peer
	SrcMask  uint8  `json:"SrcMask"`  // 44		src_mask	Source address prefix mask bits
	DstMask  uint8  `json:"DstMask"`  // 45		dst_mask	Destination address prefix mask bits
	// 46-47	pad2	Unused (zero) bytes
}

// NetFlowV5 is representing single NetFlow transmission
type NetFlowV5 struct {
	Header  netFlowV5Header   `json:"Header"`
	Records []netFlowV5Record `json:"Records"`
	Raw     string            `json:"Raw"`
}

func main() {
	ServerAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:9995")
	ServerConn, _ := net.ListenUDP("udp", ServerAddr)
	defer ServerConn.Close()

	for {
		buf := make([]byte, 4096)
		n, _, _ := ServerConn.ReadFromUDP(buf)
		go handleFlow(buf, n)
	}
}

func getTimestamp(sec []byte, nsec []byte) string {
	s := int64(sec[0])*256*256*256 +
		int64(sec[1])*256*256 +
		int64(sec[2])*256 +
		int64(sec[3])
	n := int64(nsec[0])*256*256*256 +
		int64(nsec[1])*256*256 +
		int64(nsec[2])*256 +
		int64(nsec[3])
	t := time.Unix(s, n)
	return t.Format("2006-01-02 15:04:05.000000000")
}

func handleFlow(buf []byte, n int) {
	if n < 24 {
		panic("malformed record")
	}

	var flow NetFlowV5

	//	Decode NetFlow v5 header
	var header netFlowV5Header
	header.Version = uint16(buf[0])*256 + uint16(buf[1]) //	NetFlow export format version number
	header.Count = uint16(buf[2])*256 + uint16(buf[3])   //	Number of flows exported in this packet (1-30)
	header.SysUptime = int32(buf[4])*256*256*256 +       //	Current time in milliseconds since the export device booted
		int32(buf[5])*256*256 +
		int32(buf[6])*256 +
		int32(buf[7])
	header.Timestamp = getTimestamp(
		buf[8:12],  //	Current count of seconds since 0000 UTC 1970
		buf[12:16]) //	Residual nanoseconds since 0000 UTC 1970
	header.FlowSequence = int32(buf[16])*256*256*256 + //	Sequence counter of total flows seen
		int32(buf[17])*256*256 +
		int32(buf[18])*256 +
		int32(buf[19])
	header.EngineType = uint8(buf[20])                          //	Type of flow-switching engine
	header.EngineID = uint8(buf[21])                            //	Slot number of the flow-switching engine
	header.SamplingInterval = uint16(buf[22]) + uint16(buf[24]) //	First two bits hold the sampling mode; remaining 14 bits hold value of sampling interval

	//	Decode NetFlow v5 records
	recLen := uint16(48)
	for c := uint16(0); c < header.Count; c++ {
		recBuf := buf[c*recLen+24 : c*recLen+24+recLen]
		var record netFlowV5Record
		// 00-03	srcaddr	Source IP address
		record.SrcAddr = strconv.Itoa(int(recBuf[0])) + "." +
			strconv.Itoa(int(recBuf[1])) + "." +
			strconv.Itoa(int(recBuf[2])) + "." +
			strconv.Itoa(int(recBuf[3]))

		// 04-07	dstaddr	Destination IP address
		record.DstAddr = strconv.Itoa(int(recBuf[4])) + "." +
			strconv.Itoa(int(recBuf[5])) + "." +
			strconv.Itoa(int(recBuf[6])) + "." +
			strconv.Itoa(int(recBuf[7]))

		// 08-11	nexthop	IP address of next hop router
		record.NextHop = strconv.Itoa(int(recBuf[8])) + "." +
			strconv.Itoa(int(recBuf[9])) + "." +
			strconv.Itoa(int(recBuf[10])) + "." +
			strconv.Itoa(int(recBuf[11]))

		// 12-13	input	SNMP index of input interface
		record.Input = uint16(recBuf[12])*256 + uint16(recBuf[13])

		// 14-15	output	SNMP index of output interface
		record.Output = uint16(recBuf[14])*256 + uint16(recBuf[15])

		// 16-19	dPkts	Packets in the flow
		record.DPkts = uint32(recBuf[16])*256*256*256 +
			uint32(recBuf[17])*256*256 +
			uint32(recBuf[18])*256 +
			uint32(recBuf[19])
		// record.DOctets uint32 `json:"DOctets"` // 20-23	dOctets	Total number of Layer 3 bytes in the packets of the flow
		// record.First   uint32 `json:"First"`   // 24-27	first	SysUptime at start of flow
		// record.Last    uint32 `json:"Last"`    // 28-31	last	SysUptime at the time the last packet of the flow was received

		// 32-33	srcport		TCP/UDP source port number or equivalent
		record.SrcPort = uint16(recBuf[32])*256 + uint16(recBuf[33])

		// 34-35	dstport		TCP/UDP destination port number or equivalent
		record.DstPort = uint16(recBuf[34])*256 + uint16(recBuf[35])

		// 37		tcp_flags	Cumulative OR of TCP flags
		record.TCPFlags = uint8(recBuf[37])

		// 38		prot		IP protocol type (for example, TCP = 6; UDP = 17)
		record.Prot = uint8(recBuf[38])

		// 39		tos			IP type of service (ToS)
		record.Tos = uint8(recBuf[39])

		// 40-41	src_as		Autonomous system number of the source, either origin or peer
		record.SrcAs = uint16(recBuf[40])*256 + uint16(recBuf[41])

		// 42-43	dst_as		Autonomous system number of the destination, either origin or peer
		record.DstAs = uint16(recBuf[42])*256 + uint16(recBuf[43])

		// 44		src_mask	Source address prefix mask bits
		record.SrcMask = uint8(recBuf[44])

		// 45		dst_mask	Destination address prefix mask bits
		record.DstMask = uint8(recBuf[45])

		// store decoded reckord
		flow.Records = append(flow.Records, record)
	}

	flow.Header = header
	fmt.Printf("%+v", flow)
}
