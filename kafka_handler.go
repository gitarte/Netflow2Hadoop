package main

import (
	"fmt"

	"github.com/Shopify/sarama"
)

// SendingToKafka -
func SendingToKafka(jsonFlowChanel chan string) {
	defer RecoverAnyPanic("SendingToKafka")

	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	saramaConfig.Producer.Retry.Max = 5
	saramaConfig.Producer.Return.Successes = true
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
			panic(err)
		}
		fmt.Printf("Message is stored in topic(%s)/partition(%d)/offset(%d)\n", Config.Output.Kafka.Topic, partition, offset)
	}
}
