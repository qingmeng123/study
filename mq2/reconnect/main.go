package main

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"time"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Printf("%s: %s\n", msg, err)
	}
}

func Consume() {
	defer func() {
		if err := recover(); err != nil {
			time.Sleep(3 * time.Second)
			failOnError(err.(error), "waiting 3s")
			Consume()
		}
	}()
	conn, err := amqp.Dial("amqp://guest:guest@127.0.0.1:5672/")
	failOnError(err, "Failed to dial")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.Qos(
		1,
		0,
		false,
	)
	failOnError(err, "Failed to qos")
	closeChan := make(chan *amqp.Error, 1)
	ch.NotifyClose(closeChan) //一旦消费者的channel有错误，产生一个amqp.Error，channel监听并捕捉到这个错误
	msgs, err := ch.Consume(
		"hello",
		"",
		false,
		false,
		false,
		false,
		nil)

	for {
		select {
		case e := <-closeChan:
			failOnError(e, "Failed to channel")
			close(closeChan)
			time.Sleep(5 * time.Second)
			log.Println("waiting 5s")
			return
		case msg := <-msgs:
			log.Printf("Received a message: %s", msg.Body)
			time.Sleep(time.Second * 10) //ack设置为10s，模拟超时
			err := msg.Ack(false)
			failOnError(err, "Failed to ack")
		}
	}
}

func main() {
	go Consume()
	select {}
}
