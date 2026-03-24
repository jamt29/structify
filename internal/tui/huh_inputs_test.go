package tui

import "testing"

func TestHeapStringDistinctPointers(t *testing.T) {
	p1 := heapString("a")
	p2 := heapString("b")
	if p1 == p2 {
		t.Fatal("heapString must return distinct pointers")
	}
	*p1 = "changed"
	if *p2 != "b" {
		t.Fatalf("mutating one pointer affected another: p2=%q", *p2)
	}
}
