package main

import (
	"fmt"
	"testing"

	. "github.com/arborchat/arbor-go"
	"github.com/onsi/gomega"
)

// TestRecentListConstructor ensures that invalid constructor parameters produce
// errors, but that valid sizes produce working RecentLists.
func TestRecentListConstructor(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r, err := NewRecents(-1)
	g.Expect(err).ToNot(gomega.BeNil())
	g.Expect(r).To(gomega.BeNil())

	r, err = NewRecents(0)
	g.Expect(err).ToNot(gomega.BeNil())
	g.Expect(r).To(gomega.BeNil())

	r, err = NewRecents(1)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(r).ToNot(gomega.BeNil())

	data := r.Data()
	g.Expect(data).To(gomega.BeEmpty())
}

// TestRecemtListRemoveParentVacancy tests that a parent message is replaced when
// a its child is added to the RecentList, before the queue is full
func TestRecentListRemoveParentVacancy(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	r, err := NewRecents(5)
	if err != nil {
		t.Skip("Failed to create RecentList", err)
	}
	g.Expect(r.Data()).Should(gomega.BeEmpty())

	m0, err := NewChatMessage("message 0")
	if err != nil {
		t.Skip("Failed to create message")
	}
	m0.UUID = "first"
	r.Add(m0)

	m1, err := m0.Reply("message1")
	if err != nil {
		t.Skip("Failed to reply to message")
	}
	m1.UUID = "second"

	for i := 0; i < 7; i++ {
		r.Add(m1)
		g.Expect(r.Data()).ShouldNot(gomega.ContainElement(m0.UUID))
		g.Expect(r.Data()).Should(gomega.ContainElement(m1.UUID))
		m0 = m1
		m1, err = m0.Reply("new message")
		if err != nil {
			t.Skip("Failed to reply to message")
		}
		m1.UUID = fmt.Sprintf("%dth", i+3)
	}
}

// TestRecemtListRemoveParentFull tests that a parent message is replaced when
// a its child is added to the RecentList, after the queue is full
func TestRecentListRemoveParentFull(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	r, err := NewRecents(5)
	if err != nil {
		t.Skip("Failed to create RecentList", err)
	}

	// Fill the queue with root messages
	for i := 0; i < 5; i++ {
		m1, err := NewChatMessage("new message")
		if err != nil {
			t.Skip("Failed to reply to message")
		}
		m1.AssignID()
		r.Add(m1)
	}

	// Create new root message for replying
	m0, err := NewChatMessage("message 0")
	if err != nil {
		t.Skip("Failed to create message")
	}
	m0.AssignID()
	r.Add(m0)

	// First reply
	m1, err := m0.Reply("message1")
	if err != nil {
		t.Skip("Failed to reply to message")
	}
	m1.AssignID()

	// Spin up a bunch of replies
	for i := 0; i < 7; i++ {
		r.Add(m1)
		g.Expect(r.Data()).ShouldNot(gomega.ContainElement(m0.UUID))
		g.Expect(r.Data()).Should(gomega.ContainElement(m1.UUID))
		// Shift messages down the queue
		m0 = m1
		m1, err = m0.Reply("new message")
		if err != nil {
			t.Skip("Failed to reply to message")
		}
		m1.AssignID()
	}
}

// TestRecentListAddsNewMessages ensure that before and after the queue is
// full, the newest message is ALWAYS in the list of recents, and that it
// always replaces the oldest message if the parent is not in the RecentList.
func TestRecentListAddsNewMessages(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	r, err := NewRecents(5)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(r).ToNot(gomega.BeNil())

	m0, _ := NewChatMessage("message 0")
	m0.AssignID()
	r.Add(m0)
	g.Expect(r.Data()).Should(gomega.ContainElement(m0.UUID))

	m1, _ := NewChatMessage("message 1")
	m1.AssignID()
	r.Add(m1)
	g.Expect(r.Data()).Should(gomega.ContainElement(m1.UUID))

	m2, _ := NewChatMessage("message 2")
	m2.AssignID()
	r.Add(m2)
	g.Expect(r.Data()).Should(gomega.ContainElement(m2.UUID))

	m3, _ := NewChatMessage("message 3")
	m3.AssignID()
	r.Add(m3)
	g.Expect(r.Data()).Should(gomega.ContainElement(m3.UUID))

	m4, _ := NewChatMessage("message 4")
	m4.AssignID()
	r.Add(m4)
	g.Expect(r.Data()).Should(gomega.ContainElement(m4.UUID))

	m5, _ := NewChatMessage("message 5")
	m5.AssignID()
	r.Add(m5)
	g.Expect(r.Data()).Should(gomega.ContainElement(m5.UUID))
	g.Expect(r.Data()).Should(gomega.ContainElement(m1.UUID))
	g.Expect(r.Data()).ShouldNot(gomega.ContainElement(m0.UUID))

	m6, _ := NewChatMessage("message 6")
	m6.AssignID()
	r.Add(m6)
	g.Expect(r.Data()).Should(gomega.ContainElement(m6.UUID))
	g.Expect(r.Data()).Should(gomega.ContainElement(m2.UUID))
	g.Expect(r.Data()).ShouldNot(gomega.ContainElement(m1.UUID))

	m7, _ := NewChatMessage("message 7")
	m7.AssignID()
	r.Add(m7)
	g.Expect(r.Data()).Should(gomega.ContainElement(m7.UUID))
	g.Expect(r.Data()).Should(gomega.ContainElement(m3.UUID))
	g.Expect(r.Data()).ShouldNot(gomega.ContainElement(m2.UUID))

	m8, _ := NewChatMessage("message 8")
	m8.AssignID()
	r.Add(m8)
	g.Expect(r.Data()).Should(gomega.ContainElement(m8.UUID))
	g.Expect(r.Data()).Should(gomega.ContainElement(m4.UUID))
	g.Expect(r.Data()).ShouldNot(gomega.ContainElement(m3.UUID))
}

// TestWorstCase tests that new messages are being inserted at the head of the
// queue, even when they replace their parents.
func TestWorstCase(t *testing.T) {
	// create a list
	g := gomega.NewGomegaWithT(t)
	r, err := NewRecents(3)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(r).ToNot(gomega.BeNil())

	m0, _ := NewChatMessage("message 0")
	m0.AssignID()
	r.Add(m0)

	// fill the list
	m1, _ := NewChatMessage("message 1")
	m1.AssignID()
	r.Add(m1)
	m2, _ := NewChatMessage("message 2")
	m2.AssignID()
	r.Add(m2)

	// reply to the oldest message still in the list
	m3, _ := m0.Reply("message 3")
	m3.AssignID()
	r.Add(m3)
	g.Expect(r.Data()).Should(gomega.ContainElement(m3.UUID))
	g.Expect(r.Data()).ShouldNot(gomega.ContainElement(m0.UUID))

	// add a new message to the list
	m4, _ := NewChatMessage("message 4")
	m4.AssignID()
	r.Add(m4)
	g.Expect(r.Data()).ShouldNot(gomega.ContainElement(m1.UUID))

	g.Expect(r.Data()).Should(gomega.ContainElement(m2.UUID))
	g.Expect(r.Data()).Should(gomega.ContainElement(m3.UUID))
	g.Expect(r.Data()).Should(gomega.ContainElement(m4.UUID))
}
