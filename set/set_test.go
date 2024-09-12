package set_test

import (
	"testing"

	"github.com/yaoshangnetwork/gobase/set"
)

func TestSetAdd(t *testing.T) {
	s := set.NewSet[int]()
	s.Add(1)
	if !s.Contains(1) {
		t.Error("add error")
	}
}

func TestSetRemove(t *testing.T) {
	s := set.NewSet[int]()
	s.Add(1)
	s.Remove(1)
	if s.Contains(1) {
		t.Error("remove error")
	}
}

func TestSize(t *testing.T) {
	s := set.NewSet[int]()
	s.Add(1)
	s.Add(2)
	s.Add(3)
	if s.Size() != 3 {
		t.Error("size error")
	}
}

func TestClear(t *testing.T) {
	s := set.NewSet[int]()
	s.Add(1)
	s.Add(2)
	s.Add(3)
	s.Clear()
	if s.Size() != 0 {
		t.Error("clear error")
	}
}
