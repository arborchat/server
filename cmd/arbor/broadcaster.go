package main

import (
	"log"

	"github.com/arborchat/arbor-go"
)

// Broadcaster manages sending protocol messages to all members of a pool of clients.
type Broadcaster struct {
	send       chan *arbor.ProtocolMessage
	disconnect chan arbor.Writer
	connect    chan arbor.Writer
	clients    map[arbor.Writer]struct{}
}

// NewBroadcaster creates a broadcaster.
func NewBroadcaster() *Broadcaster {
	b := &Broadcaster{
		send:       make(chan *arbor.ProtocolMessage),
		connect:    make(chan arbor.Writer),
		disconnect: make(chan arbor.Writer),
		clients:    make(map[arbor.Writer]struct{}),
	}
	go b.dispatch()
	return b
}

// dispatch runs in its own goroutine and listens for activity on all of the Broadcaster's channels.
// It updates its state when needed.
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

// Send sends a message to all clients in the managed pool of clients. Any client that
// produces an error will be automatically removed from the pool.
func (b *Broadcaster) Send(message *arbor.ProtocolMessage) {
	b.send <- message
}

// trySend attempts to send a message to the given client. If it produces an error, it
// requests that the client be removed from the managed pool.
func (b *Broadcaster) trySend(message *arbor.ProtocolMessage, client arbor.Writer) {
	log.Println("trying send to ", client)
	err := client.Write(message)
	if err != nil {
		log.Println("Error sending to client, removing: ", err)
		b.disconnect <- client
	}
}

// Add inserts a client into the managed pool.
func (b *Broadcaster) Add(client arbor.Writer) {
	b.connect <- client
}
