# netflow-collector

### Generate your dummy netflow stream
Use ```fprobe``` like this 
```bash
/usr/sbin/fprobe -iwlp3s0 -fip localhost:9995
```
To generate a flood of netflows you could start a bunch of fprobe's in a loop
```bash
for i in {1..500}
do
    /usr/sbin/fprobe -iwlp3s0 -fip 192.168.43.40:9995
done
```
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
artflow [listen address] [listen port] [the size of accumulator] [destination directory]
```
For example
```bash
artflow localhost 9995 100000 ./
```
### Quickstart for kafka
```bash
bin/zookeeper-server-start.sh config/zookeeper.properties &
bin/kafka-server-start.sh config/server.properties &

bin/kafka-topics.sh --create --zookeeper 192.168.43.20:2181 --replication-factor 1 --partitions 1 --topic test
bin/kafka-topics.sh --list --zookeeper 192.168.43.20:2181

bin/kafka-console-producer.sh --broker-list 192.168.43.20:9092 --topic test
bin/kafka-console-consumer.sh --bootstrap-server 192.168.43.20:9092 --topic test --from-beginning
```