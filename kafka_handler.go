package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"

	"github.com/Shopify/sarama"
)

// SendingToKafka -
func SendingToKafka(jsonFlowChanel chan string) {
	defer RecoverAnyPanic("SendingToKafka")

	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	saramaConfig.Producer.Retry.Max = 5
	saramaConfig.Producer.Return.Successes = true

	if Config.Output.Kafka.TLS.Enabled {
		saramaConfig.Net.TLS.Config = createTLSConfiguration()
		saramaConfig.Net.TLS.Enable = true
	}

	producer, err := sarama.NewSyncProducer(Config.Output.Kafka.BrokerList, saramaConfig)
	if err != nil {
		ExitOnError("SendingToKafka", err)
	}
	defer func() {
		if err := producer.Close(); err != nil {
			ExitOnError("SendingToKafka", err)
		}
	}()

	for {
		msg := &sarama.ProducerMessage{
			Topic: Config.Output.Kafka.Topic,
			Value: sarama.StringEncoder(<-jsonFlowChanel),
		}
		partition, offset, err := producer.SendMessage(msg)
		if err != nil {
			LogOnError("SendingToKafka", err)
		}
		if offset == 0 {
			log.Printf("First flow stored in topic(%s)/partition(%d)/offset(%d)\n", Config.Output.Kafka.Topic, partition, offset)
		} else if offset%100 == 0 {
			log.Printf("Another 100 flows stored in topic(%s)/partition(%d)/offset(%d)\n", Config.Output.Kafka.Topic, partition, offset)
		}
	}
}

func createTLSConfiguration() (t *tls.Config) {
	cert, err := tls.LoadX509KeyPair(
		Config.Output.Kafka.TLS.CertFilePath,
		Config.Output.Kafka.TLS.KeyFilePath)

	if err != nil {
		ExitOnError("createTLSConfiguration", err)
	}

	caCert, err := ioutil.ReadFile(Config.Output.Kafka.TLS.CAFilePath)
	if err != nil {
		ExitOnError("createTLSConfiguration", err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	t = &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            caCertPool,
		InsecureSkipVerify: false,
	}
	return t
}
