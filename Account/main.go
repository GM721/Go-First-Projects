package main

import (
	"Account/kafka"
	"Account/server"
	"Account/storage"
	"fmt"
)

const (
	host     = "172.16.238.3"
	port     = 5431
	user     = "user"
	password = "password"
	dbname   = "account_db"

	kafkaPort        = 5456
	kafkaGroup       = "account-group"
	topicOrder       = "order"
	topicOrderStatus = "order-status"
)

func main() {

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	storageAccount := storage.New(psqlInfo)
	defer storageAccount.DB.Close()

	var kafkaConnection = &kafka.KafkaAccountConnection{
		Producer: nil,
		Consumer: nil,
	}

	kafkaConnection.SetProducer(kafkaPort)
	kafkaConnection.SetConsumer(kafkaPort, kafkaGroup)
	kafkaConnection.SetStorage(storageAccount)
	go kafkaConnection.Listen(topicOrder, topicOrderStatus)
	defer kafkaConnection.Close()

	serv := &server.Server{}
	serv.New(storageAccount)

}
