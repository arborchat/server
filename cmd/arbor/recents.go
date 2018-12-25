package main

import (
	"fmt"
	"log"

	messages "github.com/arborchat/arbor-go"
)

// The RecentList structure is designed to be completely threadsafe
// by ensuring that all operations that touch its data occur in the
// same goroutine (dispatch()). This goroutine is launched in its
// constructor, and it just infinitely loops in the dispatch method
// waiting for activity on channels.
type RecentList struct {
	recents []string
	index   int
	full    bool
	add     chan *messages.ChatMessage
	reqData chan struct{}
	data    chan []string
}

// NewRecents takes in the number of messages requested
// and retruns a populated/populating RecentList struct.
func NewRecents(size int) (*RecentList, error) {
	if size <= 0 {
		return nil, fmt.Errorf("Invalid size for recents: %d", size)
	}
	r := &RecentList{
		recents: make([]string, 0, size),
		add:     make(chan *messages.ChatMessage),
		reqData: make(chan struct{}),
		data:    make(chan []string),
		full:    false,
		index:   0,
	}
	go r.dispatch()
	return r, nil
}

// dispatch is run as a goroutine to control access to a RecentList.
// It waits for `Add` or `Data` to be called and handles opperations
// in the order they appear in the channels r.add and r.reqData.
// This is done to keep the RecentList struct threadsafe.
func (r *RecentList) dispatch() {
	for {
		select {
		// Add function called
		case msg := <-r.add:
			log.Printf("Recents before Dispatch: %s\n", r.recents)
			log.Printf("index:   %d\n", r.index)
			log.Printf("length:  %d\n", len(r.recents))
			log.Printf("recents: %s\n", r.recents)
			parentRemoved := r.removeParent(msg)
			if !parentRemoved && len(r.recents) < cap(r.recents) {
				// Resize slice
				r.recents = r.recents[:len(r.recents)+1]
			}

			id := msg.UUID
			log.Printf("Accessing:  %d\n", r.index)
			r.recents[r.index] = id
			r.index++

			if !r.full && r.index == cap(r.recents) {
				r.full = true
			}
			log.Printf("Recents after Dispatch: %s\n", r.recents)
			r.index %= cap(r.recents)

		// Data method called
		case <-r.reqData:
			buflen := len(r.recents)
			if r.full {
				buflen = len(r.recents)
			}
			res := make([]string, buflen)
			copy(res, r.recents)
			r.data <- res
		}
	}
}

// Add attempts an addition to Recents List by sending the input ID to
// the RecentList's add channel, triggering the corrosponding selection
// in the dispatch goroutine.
func (r *RecentList) Add(msg *messages.ChatMessage) {
	r.add <- msg
}

// The Data method requests a copy of the recentlist's data by sending an
// empty value on a struct. This channel activity triggers the second case
// of the select in dispatch.
func (r *RecentList) Data() []string {
	r.reqData <- struct{}{}
	return <-r.data
}

// RemoveParent removes the parent UUID of msg, while maintaining the FIFO
// nature of the RecentList queue.
func (r *RecentList) removeParent(msg *messages.ChatMessage) bool {
	log.Printf("Adding %s\n", msg.UUID)
	log.Printf("Recents before add: %s\n", r.recents)
	if len(r.recents) == 0 {
		return false
	}

	parentID := msg.Parent
	parentIndex := -1

	// Locate parent
	for i := 0; parentIndex < 0 && i < len(r.recents); i++ {
		if r.recents[i] == parentID {
			parentIndex = i
		}
	}

	// Remove parent
	if parentIndex >= 0 {
		// Move elements
		for i := 0; len(r.recents) > 1 && (i+parentIndex)%len(r.recents) != r.index; i++ {
			index := (i + parentIndex) % len(r.recents)
			log.Printf("index: %d\n", i)
			log.Println()
			r.recents[index] = r.recents[(index+1)%len(r.recents)]
		}
		r.index--
	}
	log.Printf("Recents after add: %s\n", r.recents)
	return parentIndex >= 0
}
