package main

import (
	"flag"
	"log"
	"net"

	"github.com/arborchat/arbor-go"
)

func main() {
	ruser := flag.String("ruser", "root", "The username of the root message")
	rid := flag.String("rid", "", "The id of the root message")
	rcontent := flag.String("rcontent", "Welcome to our server!", "The content of the root message")
	recentSize := flag.Int("recent-size", 100, "The number of messages to keep in the recents list")
	flag.Parse()
	messages := arbor.NewStore()
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
	m, err := arbor.NewChatMessage(*rcontent)
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
	toWelcome := make(chan arbor.Writer)
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
			continue
		}
		rw, err := arbor.NewProtocolReadWriter(conn)
		if err != nil {
			log.Println(err)
			continue
		}
		broadcaster.Add(rw)
		go handleClient(rw, recents, messages, broadcaster)
		toWelcome <- rw
	}
}

// handleWelcomes reads the toWelcome channel and sends a WELCOME message to each client that
// it receives from that channel.
func handleWelcomes(rootId string, recents *RecentList, toWelcome chan arbor.Writer) {
	for client := range toWelcome {
		msg := arbor.ProtocolMessage{
			Type:  arbor.WelcomeType,
			Root:  rootId,
			Major: 0,
			Minor: 1,
		}
		msg.Recent = recents.Data()

		err := client.Write(&msg)
		if err != nil {
			log.Println(err)
			continue
		}
		log.Println("Welcome message: ", msg.String())

	}
}

// handleClient is a goroutine that is instantiated for every client. It reads from the client and launches handlers
// for every message until it encounters an error. All errors make it terminate the connection to the client.
func handleClient(client arbor.ReadWriteCloser, recents *RecentList, store *arbor.Store, broadcaster *Broadcaster) {
	defer client.Close()
	for {
		message := new(arbor.ProtocolMessage)
		err := client.Read(message)
		if err != nil {
			log.Println(err)
			return
		}
		switch message.Type {
		case arbor.QueryType:
			log.Println("Handling query for " + message.ChatMessage.UUID)
			go handleQuery(message, client, store)
		case arbor.NewMessageType:
			go handleNewMessage(message, recents, store, broadcaster)
		default:
			log.Println("Unrecognized message type", message.Type)
			return
		}
	}
}

// handleQuery responds to a query if possible.
func handleQuery(msg *arbor.ProtocolMessage, out arbor.Writer, store *arbor.Store) {
	result := store.Get(msg.ChatMessage.UUID)
	if result == nil {
		log.Println("Unable to find queried id: " + msg.ChatMessage.UUID)
		return
	}
	msg.ChatMessage = result
	msg.Type = arbor.NewMessageType
	err := out.Write(msg)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("Query response: ", msg.String())
}

// handleNewMessage updates the in-memory store of messages, updates the recents list, and broadcasts the new message.
func handleNewMessage(msg *arbor.ProtocolMessage, recents *RecentList, store *arbor.Store, broadcaster *Broadcaster) {
	err := msg.ChatMessage.AssignID()
	if err != nil {
		log.Println("Error creating new message", err)
	}
	recents.Add(msg.ChatMessage)
	store.Add(msg.ChatMessage)
	broadcaster.Send(msg)
}
