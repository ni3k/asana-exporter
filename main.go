package main

import (
	asanaexporter "asana-exporter/asana"
	"asana-exporter/rabbitmq"
	"fmt"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

func main() {
	rmqInstance, err := rabbitmq.NewRabbitMq("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal(err)
	}
	defer rmqInstance.Close()

	q, err := rmqInstance.InitQueue("exports")
	if err != nil {
		log.Fatal(err)
	}

	// start an initial export
	fmt.Println("requesting export for users")
	rmqInstance.PublishToChannel(string(asanaexporter.UsersEntity), q.Name)
	fmt.Println("requesting export for projects")
	rmqInstance.PublishToChannel(string(asanaexporter.ProjectsEntity), q.Name)

	done := make(chan bool)
	go func(q amqp091.Queue) {
		timer1 := time.NewTicker(5 * time.Minute)
		timer2 := time.NewTicker(30 * time.Second)
		for {
			select {
			case <-timer1.C:
				fmt.Println("requesting export for users")
				rmqInstance.PublishToChannel(string(asanaexporter.UsersEntity), q.Name)
				fmt.Println("requesting export for projects")
				rmqInstance.PublishToChannel(string(asanaexporter.ProjectsEntity), q.Name)
			case <-timer2.C:
				fmt.Println("requesting export for users")
				rmqInstance.PublishToChannel(string(asanaexporter.UsersEntity), q.Name)
				fmt.Println("requesting export for projects")
				rmqInstance.PublishToChannel(string(asanaexporter.ProjectsEntity), q.Name)
			}
		}
	}(q)
	<-done
}
