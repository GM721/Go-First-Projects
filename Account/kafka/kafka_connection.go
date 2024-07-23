package kafka

import (
	"Account/entities"
	"Account/storage"
	"encoding/json"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"os"
	"strconv"
)

type KafkaAccountConnection struct {
	Producer *kafka.Producer
	Consumer *kafka.Consumer

	storage *storage.StorageAccount
}

func (kafkaConnection *KafkaAccountConnection) SetProducer(port int) {

	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": fmt.Sprintf("172.16.238.4:%d", port),
		"acks":              "all"})

	if err != nil {
		fmt.Printf("Failed to create Producer: %s", err)
		os.Exit(1)
	}

	go func() {
		for e := range producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Failed to deliver message: %v\n", ev.TopicPartition)
				} else {
					fmt.Printf("Produced event to TopicResponseOrder %s: key = %-10s value = %s\n",
						*ev.TopicPartition.Topic, string(ev.Key), string(ev.Value))
				}
			}
		}
	}()

	kafkaConnection.Producer = producer
}

func (kafkaConnection *KafkaAccountConnection) SetConsumer(port int, group string) {

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": fmt.Sprintf("172.16.238.4:%d", port),
		"group.id":          group,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		panic(err)
	}

	kafkaConnection.Consumer = consumer
}

func (kafkaConnection *KafkaAccountConnection) SetStorage(storage *storage.StorageAccount) {
	kafkaConnection.storage = storage
}

func (kafkaConnection *KafkaAccountConnection) SendMessage(key uint, message string, topic string) {

	var err = kafkaConnection.Producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            []byte(strconv.Itoa(int(key))),
		Value:          []byte(message),
	}, nil)
	if err != nil {
		fmt.Println(err)
	}

}

func (kafkaConnection *KafkaAccountConnection) Listen(topicGet string, topicResponse string) {

	err := kafkaConnection.Consumer.SubscribeTopics([]string{topicGet}, nil)
	if err != nil {
		panic(err)
	}

	for {
		msg, err := kafkaConnection.Consumer.ReadMessage(-1)
		if err != nil {
			fmt.Println(err)
			continue
		}

		var order entities.Order
		err = json.Unmarshal(msg.Value, &order)
		if err != nil {
			fmt.Println(err)
			continue
		}

		deposit, err := kafkaConnection.storage.GetDeposit(order.ClientID)
		if err != nil || deposit < order.Amount {
			fmt.Println(555)
			kafkaConnection.SendMessage(order.ID, "CANCELED", topicResponse)
			continue
		}

		kafkaConnection.storage.MU.Lock()
		ok := kafkaConnection.storage.DecreaseDeposit(order.ClientID, order.Amount)
		kafkaConnection.storage.MU.Unlock()
		if !ok {
			kafkaConnection.SendMessage(order.ID, "CANCELED", topicResponse)
			continue
		}
		kafkaConnection.SendMessage(order.ID, "PAID", topicResponse)
	}
}

func (kafkaConnection *KafkaAccountConnection) Close() {

	kafkaConnection.Producer.Close()

	err := kafkaConnection.Consumer.Close()
	if err != nil {
		panic(err)
	}
}

func i32tob(val uint32) []byte {
	r := make([]byte, 4)
	for i := uint32(0); i < 4; i++ {
		r[i] = byte((val >> (8 * i)) & 0xff)
	}
	return r
}
