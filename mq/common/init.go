package common

import (
	"github.com/streadway/amqp"
	"log"
)

//var Ch *amqp.Channel

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s:%s", msg, err)
	}

}

func MustInit() (*amqp.Connection, *amqp.Channel) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")

	err = ch.ExchangeDeclare(
		"logs",   // name
		"fanout", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	return conn, ch
}
