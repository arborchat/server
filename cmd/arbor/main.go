package main

import (
	"log"
	"net"
	"os"

	. "github.com/arborchat/arbor-go"
)

func main() {
	messages := NewStore()
	broadcaster := NewBroadcaster()
	address := ":7777"
	recents, err := NewRecents(10)
	if err != nil {
		log.Fatalln("Unable to initialize Recents", err)
	}
	//serve
	if len(os.Args) > 1 {
		address = os.Args[1]
	}
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Server listening on", address)
	m, err := NewChatMessage("Welcome to our server!")
	if err != nil {
		log.Println(err)
	}
	m.Username = "root"
	err = m.AssignID()
	if err != nil {
		log.Println(err)
	}
	messages.Add(m)
	toWelcome := make(chan chan<- *ProtocolMessage)
	go handleWelcomes(m.UUID, recents, toWelcome)
	log.Println("Root message UUID is " + m.UUID)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
		}
		fromClient := MakeMessageReader(conn)
		toClient := MakeMessageWriter(conn)
		broadcaster.Add(toClient)
		go handleClient(fromClient, toClient, recents, messages, broadcaster)
		toWelcome <- toClient
	}
}

func handleWelcomes(rootId string, recents *RecentList, toWelcome chan chan<- *ProtocolMessage) {
	for client := range toWelcome {
		msg := ProtocolMessage{
			Type:  WelcomeType,
			Root:  rootId,
			Major: 0,
			Minor: 1,
		}
		msg.Recent = recents.Data()

		client <- &msg
		log.Println("Welcome message: ", msg.String())

	}
}

func handleClient(from <-chan *ProtocolMessage, to chan<- *ProtocolMessage, recents *RecentList, store *Store, broadcaster *Broadcaster) {
	for message := range from {
		switch message.Type {
		case QueryType:
			log.Println("Handling query for " + message.ChatMessage.UUID)
			go handleQuery(message, to, store)
		case NewMessageType:
			go handleNewMessage(message, recents, store, broadcaster)
		default:
			log.Println("Unrecognized message type", message.Type)
			continue
		}
	}
}

func handleQuery(msg *ProtocolMessage, out chan<- *ProtocolMessage, store *Store) {
	result := store.Get(msg.ChatMessage.UUID)
	if result == nil {
		log.Println("Unable to find queried id: " + msg.ChatMessage.UUID)
		return
	}
	msg.ChatMessage = result
	msg.Type = NewMessageType
	out <- msg
	log.Println("Query response: ", msg.String())
}

func handleNewMessage(msg *ProtocolMessage, recents *RecentList, store *Store, broadcaster *Broadcaster) {
	err := msg.ChatMessage.AssignID()
	if err != nil {
		log.Println("Error creating new message", err)
	}
	recents.Add(msg.ChatMessage.UUID)
	store.Add(msg.ChatMessage)
	broadcaster.Send(msg)
}
