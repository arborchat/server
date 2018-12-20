package main

import (
	"testing"

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
