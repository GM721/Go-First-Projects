package kafka

import (
	"Order/entities"
	"encoding/json"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"os"
	"strconv"
)

type KafkaOrderConnection struct {
	Producer *kafka.Producer
	Consumer *kafka.Consumer

	TopicSendOrder string

	GetMessages map[uint]chan string
}

func (kafkaConnection *KafkaOrderConnection) SetProducer(port int, topicSendOrder string) {

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
					fmt.Printf("Produced event to TopicSendOrder %s: key = %-10s value = %s\n",
						*ev.TopicPartition.Topic, string(ev.Key), string(ev.Value))
				}
			}
		}
	}()

	kafkaConnection.Producer = producer
	kafkaConnection.TopicSendOrder = topicSendOrder
}

func (kafkaConnection *KafkaOrderConnection) SetConsumer(port int, group string) {

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

func (kafkaConnection *KafkaOrderConnection) SendMessageOrder(order entities.Order) error {

	value, err := json.Marshal(order)
	if err != nil {
		return err
	}

	err = kafkaConnection.Producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &kafkaConnection.TopicSendOrder, Partition: kafka.PartitionAny},
		Value:          value,
	}, nil)
	if err != nil {
		return err
	}

	kafkaConnection.GetMessages[order.ID] = make(chan string, 1)
	return nil
}

func (kafkaConnection *KafkaOrderConnection) Listen(topic string) {

	kafkaConnection.GetMessages = make(map[uint]chan string)
	err := kafkaConnection.Consumer.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		panic(err)
	}

	for {
		msg, err := kafkaConnection.Consumer.ReadMessage(-1)
		if err != nil {
			fmt.Println(err)
			continue
		}
		key, _ := strconv.Atoi(string(msg.Key))
		kafkaConnection.GetMessages[uint(key)] <- string(msg.Value)
	}
}

func (kafkaConnection *KafkaOrderConnection) Close() {

	kafkaConnection.Producer.Close()

	err := kafkaConnection.Consumer.Close()
	if err != nil {
		panic(err)
	}
}
