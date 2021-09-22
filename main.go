package main

import (
	"fmt"
	"log"

	"github.com/Vractos/go-reports-service/infra/kafka"
	"github.com/Vractos/go-reports-service/infra/repository"
	"github.com/Vractos/go-reports-service/usacase"
	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{
			"http://host.docker.internal:9200",
		},
	})
	if err != nil {
		log.Fatal("Error connecting to elasticsearch")
	}

	repo := repository.TransactionElasticRepository{
		Client: *client,
	}

	msgChan := make(chan *ckafka.Message)
	consumer := kafka.NewKafkaConsumer(msgChan)
	go consumer.Consume()
	for msg := range msgChan {
		err := usacase.GenetareReport(msg.Value, repo)
		if err != nil {
			fmt.Println(err)
		}
	}
}
