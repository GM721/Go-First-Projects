package main

import (
	"Order/kafka"
	"Order/server"
	"Order/storage"
	"fmt"
)

const (
	host     = "172.16.238.2"
	port     = 5430
	user     = "user"
	password = "password"
	dbname   = "order_db"

	kafkaPort        = 5456
	kafkaGroup       = "order-group"
	topicOrder       = "order"
	topicOrderStatus = "order-status"
)

func main() {

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	storageOrder, err := storage.New(psqlInfo)
	if err != nil {
		panic(err)
	}
	defer storageOrder.DB.Close()

	var kafkaConnection = &kafka.KafkaOrderConnection{
		Producer:       nil,
		Consumer:       nil,
		TopicSendOrder: topicOrder,
		GetMessages:    nil,
	}

	kafkaConnection.SetProducer(kafkaPort, topicOrder)
	kafkaConnection.SetConsumer(kafkaPort, kafkaGroup)
	go kafkaConnection.Listen(topicOrderStatus)
	defer kafkaConnection.Close()

	serv := &server.Server{}
	serv.New(storageOrder, kafkaConnection)

}
