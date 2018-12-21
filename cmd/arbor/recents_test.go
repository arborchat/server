package main

import (
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

func TestRecentListRemoveParent(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	r1, _ := NewRecents(5)
	r2, _ := NewRecents(5)

	m0, _ := NewChatMessage("message 0")
	m0.AssignID()
	r1.Add(m0)
	r2.Add(m0)

	// r1 contains m1 with no parent ID
	// r2 contains m1 with m0 as parent
	// so m0 should be removed from recents
	m1, _ := NewChatMessage("message 1")
	m1.AssignID()
	r1.Add(m1)
	m1.Parent = m0.UUID
	r2.Add(m1)
	g.Expect(r1.Data()).Should(gomega.ContainElement(m0.UUID))
	g.Expect(r2.Data()).ShouldNot(gomega.ContainElement(m0.UUID))

	//m2, _ := NewChatMessage("message 2")
	//m3, _ := NewChatMessage("message 3")
	//m4, _ := NewChatMessage("message 4")
	//m5, _ := NewChatMessage("message 5")
	//m6, _ := NewChatMessage("message 6")
	//m7, _ := NewChatMessage("message 7")
	//m8, _ := NewChatMessage("message 8")
}

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
