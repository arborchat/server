package main

import (
	"fmt"

	messages "github.com/arborchat/arbor-go"
)

// LeafList stores recent identifiers for elements in a tree. It has a fixed size, and will
// remove old elements in order to add new ones once it reaches its capacity.
type LeafList struct {
	elements    []string
	insertPoint int
}

// NewLeafList creates a LeafList with the given capacity.
func NewLeafList(capacity int) (*LeafList, error) {
	if capacity < 1 {
		return nil, fmt.Errorf("Illegal capacity %d", capacity)
	}
	return &LeafList{
		elements:    make([]string, 0, capacity),
		insertPoint: 0,
	}, nil
}

// Has returns whether or not the LeafList contains the provided identifier.
func (r *LeafList) Has(s string) bool {
	for _, v := range r.elements {
		if v == s {
			return true
		}
	}
	return false
}

// Add inserts the given element into the LeafList. If the element is already present, Add will
// do nothing.
func (r *LeafList) Add(s string) {
	if r.Has(s) {
		// If `s` is already present, no reason to do any extra work.
		return
	}
	// if we are at the end of the list and are able to grow, grow
	if r.insertPoint == len(r.elements) && len(r.elements) < cap(r.elements) {
		r.elements = r.elements[:len(r.elements)+1]
	}
	// insert our new element
	r.elements[r.insertPoint] = s
	r.insertPoint = (r.insertPoint + 1) % cap(r.elements)
}

// Replace will remove `old` and then insert `s`. This is useful for removing the parent identifier
// when adding a child.
func (r *LeafList) Replace(old, s string) {
	// find where `old` is within the LeafList
	oldIndex := -1
	for index := 0; index < len(r.elements); index++ {
		if r.elements[index] == old {
			oldIndex = index
			break
		}
	}
	// we only need to shift if `old` is present and is not the element that is about to be replaced
	if oldIndex > -1 && oldIndex != r.insertPoint {
		// shift every element back one position until we reach the insertionPoint
		for oldIndex++; oldIndex != r.insertPoint; oldIndex = (oldIndex + 1) % cap(r.elements) {
			r.elements[oldIndex-1] = r.elements[oldIndex]
		}
		// ensure that the Add() we are about to perform overwrites the duplicated element that
		// we create at the end of the loop
		r.insertPoint--
	}
	r.Add(s)
}

// AddOrReplace will replace `old` if it is present in the LeafList, and will add `s` if it is not.
func (r *LeafList) AddOrReplace(old, s string) {
	if r.Has(old) {
		r.Replace(old, s)
		return
	}
	r.Add(s)
}

// Elements returns a copy of the elements contained within the LeafList.
func (r *LeafList) Elements() []string {
	out := make([]string, len(r.elements))
	copy(out, r.elements)
	return out
}

// RecentList provides threadsafe access to a list of recent message identifiers.
type RecentList struct {
	*LeafList
	add     chan *messages.ChatMessage
	reqData chan struct{}
	data    chan []string
}

// NewRecents takes in the number of messages requested
// and returns a RecentList struct.
func NewRecents(size int) (*RecentList, error) {
	leaves, err := NewLeafList(size)
	if err != nil {
		return nil, err
	}
	r := &RecentList{
		LeafList: leaves,
		add:      make(chan *messages.ChatMessage),
		reqData:  make(chan struct{}),
		data:     make(chan []string),
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
			r.AddOrReplace(msg.Parent, msg.UUID)
		// Data method called
		case <-r.reqData:
			r.data <- r.Elements()
		}
	}
}

// Add attempts to insert a message's id. It will replace its parent ID if the parent ID is present.
func (r *RecentList) Add(msg *messages.ChatMessage) {
	// Send the input ID to the add channel, triggering the corrosponding selection
	// in the dispatch goroutine.
	r.add <- msg
}

// Data requests a copy of the recentlist's data.
func (r *RecentList) Data() []string {
	// Send an empty value on a struct. This channel activity triggers the second case
	// of the select in dispatch.
	r.reqData <- struct{}{}
	return <-r.data
}
