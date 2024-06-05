package main

import (
	asanaexporter "asana-exporter/asana"
	"asana-exporter/rabbitmq"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func FailOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	rmqInstance, err := rabbitmq.NewRabbitMq("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal(err)
	}
	defer rmqInstance.Close()

	q, err := rmqInstance.InitQueue("exports")
	if err != nil {
		log.Fatal(err)
	}

	msgs, err := rmqInstance.Ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Fatal(err)
	}
	var forever chan struct{}

	exporter := asanaexporter.NewAsanaExtractor("https://app.asana.com/api/1.0/", os.Getenv("ASANA_API"))

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			entityName := string(d.Body)
			entities, err := exporter.RetrieveEntities(entityName)
			if err != nil {
				fmt.Println("Failed to retrieve entities: ", d.Body)
				continue
			}
			err = exporter.StoreEntities(entityName, entities)
			if err != nil {
				fmt.Println(err)
				fmt.Println("failed to store entities")
			}
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever

}
