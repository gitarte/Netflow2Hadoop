# NetflowCollector
This application opens UDP socket for NetFlow datagrams on given ```ListenParams```. It decodes the content of each flow into JSON shape according to ```ConfigV5Header``` and ```ConfigV5Header``` description. After successful decoding it stores ```Output.ChunkSize``` number of JSONs into file on local file system or sends it for external storage on HDFS. It automatically produces new file if ```Output.ChunkSize``` was exceeded due to the amount of incoming flows. 

If configuration says that flows has to be sent into Kafka topic, then each decoded flow is published separately without any accumulation.
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
                "Enabled"      : true,
                "CertFilePath" : "./resources/cert.crt",
                "KeyFilePath"  : "./resources/key.crt",
                "CAFilePath"   : "./resources/ca.crt"
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

I believe that other options are self explanatory.
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
artflow [path to configuration JSON file] [path to log file]
```
For example
```bash
./artflow ./resources/config.json ./
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
### Quick start for Kafka with TLS
#### On Kafka machine
Create keystore:
```bash
keytool -keystore kafka.server.keystore.jks -alias localhost -genkey
```
Create your own Certificate Authority (CA):
```bash
openssl req -new -x509 -keyout ca-key -out ca-cert -days 3650
```
Add the generated CA to truststore:
```bash
keytool -keystore kafka.server.truststore.jks -alias CARoot -import -file ca-cert
```
Export the certificate from the keystore:
```bash
keytool -keystore kafka.server.keystore.jks -alias localhost -certreq -file cert-file
```
Sign it with the CA:
```bash
openssl x509 -req -CA ca-cert -CAkey ca-key -in cert-file -out cert-signed -days 3650 -CAcreateserial
```
Import both the certificate of the CA and the signed certificate into the broker keystore:
```bash
keytool -keystore kafka.server.keystore.jks -alias CARoot    -import -file ca-cert
keytool -keystore kafka.server.keystore.jks -alias localhost -import -file cert-signed
```






##### On NetflowCollector machine
