package main

import (
	"flag"
	"log"
	"net"

	. "github.com/arborchat/arbor-go"
)

func main() {
	ruser := flag.String("ruser", "root", "The username of the root message")
	rid := flag.String("rid", "", "The id of the root message")
	rcontent := flag.String("rcontent", "Welcome to our server!", "The content of the root message")
	recentSize := flag.Int("recent-size", 100, "The number of messages to keep in the recents list")
	flag.Parse()
	messages := NewStore()
	broadcaster := NewBroadcaster()
	address := ":7777"
	recents, err := NewRecents(*recentSize)
	if err != nil {
		log.Fatalln("Unable to initialize Recents", err)
	}
	//serve
	if len(flag.Args()) > 0 {
		address = flag.Arg(0)
	}
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Server listening on", address)
	m, err := NewChatMessage(*rcontent)
	if err != nil {
		log.Println(err)
	}
	m.Username = *ruser
	if *rid == "" {
		err = m.AssignID()
		if err != nil {
			log.Println(err)
		}
	} else {
		m.UUID = *rid
	}
	messages.Add(m)
	toWelcome := make(chan chan<- *ProtocolMessage)
	go handleWelcomes(m.UUID, recents, toWelcome)
	log.Println("Root message UUID is " + m.UUID)

	// Forever listen for incoming connections
	// If connection recieved, add to broadcaster list,
	// start a client connection goroutine,
	// and send welcome message
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

// handleClient is a goroutine that is instantiated for every client.
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
	recents.Add(msg.ChatMessage)
	store.Add(msg.ChatMessage)
	broadcaster.Send(msg)
}
