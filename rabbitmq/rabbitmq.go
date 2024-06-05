package rabbitmq

import (
	"context"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMq struct {
	url  string
	conn *amqp.Connection
	Ch   *amqp.Channel
}

func NewRabbitMq(url string) (*RabbitMq, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	r := RabbitMq{url: url}
	r.conn = conn
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	r.Ch = ch
	return &r, nil
}

func (r *RabbitMq) InitQueue(name string) (amqp.Queue, error) {
	q, err := r.Ch.QueueDeclare(
		name,
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	return q, err
}

func (r *RabbitMq) PublishToChannel(message string, qName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return r.Ch.PublishWithContext(ctx,
		"",    // exchange
		qName, // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})
}

func (r *RabbitMq) Close() {
	r.conn.Close()
	r.Ch.Close()
}
