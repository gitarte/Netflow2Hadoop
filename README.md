# NetflowCollector
This application opens UDP socket for NetFlow datagrams on given ```ListenParams```. It decodes the content of each flow into JSON shape according to ```ConfigV5Header``` and ```ConfigV5Header``` description. After successful decoding it stores ```Output.ChunkSize``` number of JSONs into file on local file system or sends it for external storage on HDFS. It automatically produces new file if ```Output.ChunkSize``` was exceeded due to the amount of incoming flows. If configuration says that flows has to be sent into Kafka topic, then no accumulation occurs. Instead each decoded flow is published separately.
### Configuration
This application expects command line argument that shows path to configuration JSON file of following shape:
```json
{
    "ListenParams" : "127.0.0.1:9995",
    "Output"       : {
        "ChunkSize"      : 500,
        "LocalFS"        : {
            "Enabled"    : false,
            "Path"       : "./"
        },
        "HDFS"           : {
            "Enabled"    : false,
            "Path"       : "./"
        },
        "Kafka"          : {
            "Enabled"    : true,
            "BrokerList" : [
                "192.168.43.20:9092"
            ],
            "Topic"      : "test",
            "TLS"        : {
                "Enabled"  : true,
                "CertPath" : "./resources/cert.crt"
            }
        }
    },
	"ConfigV5Header" : {
        "Version"          : false,
        "Count"            : true,
        "SysUptime"        : true,
        "Timestamp"        : true,
        "FlowSequence"     : true,
        "EngineType"       : false,
        "EngineID"         : false,
        "SamplingInterval" : false
    },
    "ConfigV5Record" : {
        "SrcAddr"  : true,
        "DstAddr"  : true,
        "NextHop"  : false,
        "Input"    : false,
        "Output"   : false,
        "DPkts"    : false,
        "DOctets"  : false,
        "First"    : false,
        "Last"     : false,
        "SrcPort"  : true,
        "DstPort"  : true,
        "TCPFlags" : true,
        "Prot"     : true,
        "Tos"      : false,
        "SrcAs"    : false,
        "DstAs"    : false,
        "SrcMask"  : false,
        "DstMask"  : false
    }
}
```
```Output.ChunkSize``` defines how many flows have to be accumulated in local memory before they will be stored to file system either local or HDFS. The output file in booth cases consists of JSON array of this size. This does not however apply to Kafka message system, where each flow is published separately as soon as it is decoded into JSON.

Boolean values in ```ConfigV5Header``` and ```ConfigV5Record``` say weather decode given NetFlow field or not. If ```false``` then the field will still be present in output but with Go's zero value. Check data types in ```netflow_v5.go```. 
### Build
For Linux
```bash
go build -o artflow *.go
```
For Windows (cross compile)
```bash
GOOS=windows GOARCH=amd64 go build -o artflow.exe *.go
```
### Start collector
```bash
artflow [path to configuration json file]
```
For example
```bash
./artflow ./resources/config.json
```
### Generate your dummy netflow stream
Use ```fprobe``` like this 
```bash
/usr/sbin/fprobe -iwlp3s0 -fip localhost:9995
```
To generate a flood of netflows you could start a bunch of fprobes in a loop
```bash
for i in {1..500}
do
    /usr/sbin/fprobe -iwlp3s0 -fip 192.168.43.40:9995
done
```
### Quick start for Kafka
```bash
bin/zookeeper-server-start.sh config/zookeeper.properties &
bin/kafka-server-start.sh config/server.properties &

bin/kafka-topics.sh --create --zookeeper 192.168.43.20:2181 --replication-factor 1 --partitions 1 --topic test
bin/kafka-topics.sh --list --zookeeper 192.168.43.20:2181

bin/kafka-console-producer.sh --broker-list 192.168.43.20:9092 --topic test
bin/kafka-console-consumer.sh --bootstrap-server 192.168.43.20:9092 --topic test --from-beginning
```