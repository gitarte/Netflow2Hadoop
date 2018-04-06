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