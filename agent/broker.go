package main

import (
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
)

type Subscriber func(msg mqtt.Message)

type Broker struct {
	client      mqtt.Client
	subscribers map[string]Subscriber
}

func NewBroker(address, username, password string) *Broker {
	b := &Broker{}

	opts := mqtt.NewClientOptions().AddBroker(fmt.Sprintf("tcp://%s", address))
	opts.SetUsername(username)
	opts.SetPassword(password)
	opts.SetAutoReconnect(true)
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		for topic, subscriber := range b.subscribers {
			if token := b.client.Subscribe(topic, 1, func(client mqtt.Client, msg mqtt.Message) {
				logrus.WithFields(logrus.Fields{"topic": topic}).Info("Received message on topic")
				subscriber(msg)
			}); token.Wait() && token.Error() != nil {
				logrus.WithFields(logrus.Fields{"topic": topic, "err": token.Error()}).Error("Failed to subscribe to topic")
			}

			logrus.WithFields(logrus.Fields{"topic": topic}).Info("Subscribed to topic")
		}
	})
	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		logrus.WithFields(logrus.Fields{"err": err}).Error("Broker connection lost")

		b.client = client
		b.Connect()
	})

	b.client = mqtt.NewClient(opts)
	b.subscribers = make(map[string]Subscriber)

	return b
}

func (b *Broker) Connect() {
	for {
		if token := b.client.Connect(); token.Wait() && token.Error() != nil {
			logrus.WithFields(logrus.Fields{"err": token.Error()}).Error("Failed to connect to broker")

			time.Sleep(time.Second)
			continue
		}

		logrus.Info("Connected to broker")

		break
	}
}

func (b *Broker) Subscribe(topic string, subscriber Subscriber) {
	b.subscribers[topic] = subscriber
}
