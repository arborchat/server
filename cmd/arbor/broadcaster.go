package main

import (
	"log"

	messages "github.com/arborchat/arbor-go"
)

type Broadcaster struct {
	send       chan *messages.ProtocolMessage
	disconnect chan chan<- *messages.ProtocolMessage
	connect    chan chan<- *messages.ProtocolMessage
	clients    map[chan<- *messages.ProtocolMessage]struct{}
}

func NewBroadcaster() *Broadcaster {
	b := &Broadcaster{
		send:       make(chan *messages.ProtocolMessage),
		connect:    make(chan chan<- *messages.ProtocolMessage),
		disconnect: make(chan chan<- *messages.ProtocolMessage),
		clients:    make(map[chan<- *messages.ProtocolMessage]struct{}),
	}
	go b.dispatch()
	return b
}

func (b *Broadcaster) dispatch() {
	for {
		select {
		case message := <-b.send:
			for client := range b.clients {
				go b.trySend(message, client)
			}
		case newclient := <-b.connect:
			b.clients[newclient] = struct{}{}

		case deadclient := <-b.disconnect:
			delete(b.clients, deadclient)
		}
	}
}

func (b *Broadcaster) Send(message *messages.ProtocolMessage) {
	b.send <- message
}

func (b *Broadcaster) trySend(message *messages.ProtocolMessage, client chan<- *messages.ProtocolMessage) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("Error sending to client, removing: ", err)
			b.disconnect <- client
		}
	}()
	log.Println("trying send to ", client)
	client <- message
}

func (b *Broadcaster) Add(client chan<- *messages.ProtocolMessage) {
	b.connect <- client
}
