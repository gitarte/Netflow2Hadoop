package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

// Configuration - configuration structure
type Configuration struct {
	ListenParams   string         `json:"ListenParams"` //	[listen address]:[listen port]
	Output         Output         `json:"Output"`       //	where to store or send collected data
	Configv5Header ConfigV5Header `json:"ConfigV5Header"`
	ConfigV5Record ConfigV5Record `json:"ConfigV5Record"`
}

// Output -
type Output struct {
	ChunkSize int         `json:"ChunkSize"` //	the size of accumulator
	LocalFS   LocalFSConf `json:"LocalFS"`   //	[listen address]:[listen port]
	HDFS      HDFSConf    `json:"HDFS"`
	Kafka     KafkaConf   `json:"Kafka"`
}

// LocalFSConf -
type LocalFSConf struct {
	Enabled bool   `json:"Enabled"`
	Path    string `json:"Path"`
}

// HDFSConf -
type HDFSConf struct {
	Enabled bool   `json:"Enabled"`
	Path    string `json:"Path"`
}

// KafkaConf -
type KafkaConf struct {
	Enabled    bool         `json:"Enabled"`
	BrokerList []string     `json:"BrokerList"` //	[listen address]:[listen port]
	Topic      string       `json:"Topic"`
	TLS        KafkaConfTLS `json:"TLS"`
}

// KafkaConfTLS -
type KafkaConfTLS struct {
	Enabled      bool   `json:"Enabled"`
	CertFilePath string `json:"CertFilePath"`
	KeyFilePath  string `json:"KeyFilePath"`
	CAFilePath   string `json:"CAFilePath"`
}

// ConfigV5Header - struct of enable/disable switches for netflow V5 header fields
type ConfigV5Header struct {
	Version          bool `json:"Version"`
	Count            bool `json:"Count"`
	SysUptime        bool `json:"SysUptime"`
	Timestamp        bool `json:"Timestamp"`
	FlowSequence     bool `json:"FlowSequence"`
	EngineType       bool `json:"EngineType"`
	EngineID         bool `json:"EngineID"`
	SamplingInterval bool `json:"SamplingInterval"`
}

// ConfigV5Record - struct of enable/disable switches for netflow V5 record fields
type ConfigV5Record struct {
	SrcAddr  bool `json:"SrcAddr"`
	DstAddr  bool `json:"DstAddr"`
	NextHop  bool `json:"NextHop"`
	Input    bool `json:"Input"`
	Output   bool `json:"Output"`
	DPkts    bool `json:"DPkts"`
	DOctets  bool `json:"DOctets"`
	First    bool `json:"First"`
	Last     bool `json:"Last"`
	SrcPort  bool `json:"SrcPort"`
	DstPort  bool `json:"DstPort"`
	TCPFlags bool `json:"TCPFlags"`
	Prot     bool `json:"Prot"`
	Tos      bool `json:"Tos"`
	SrcAs    bool `json:"SrcAs"`
	DstAs    bool `json:"DstAs"`
	SrcMask  bool `json:"SrcMask"`
	DstMask  bool `json:"DstMask"`
}

// ReadConfig - reads configuration JSON file and feeds configuration struct
func ReadConfig(c *Configuration) {
	file, e := ioutil.ReadFile(os.Args[1])
	if e != nil {
		log.Fatalf("cannot read configuration file: %v\n", e)
	}
	json.Unmarshal(file, c)
}
