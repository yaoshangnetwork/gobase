package sliceutils_test

import (
	"gobase/utils/sliceutils"
	"testing"
)

func TestFilter(t *testing.T) {
	slice := []int{1, 2, 1, 1, 1}
	filtered := sliceutils.Filter(slice, func(item int) bool { return item > 1 })
	if len(filtered) != 1 && filtered[0] != 2 {
		t.Error("filter error")
	}
}

func TestMap(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	filtered := sliceutils.Map(slice, func(item int) int { return item * 2 })
	if len(filtered) != len(slice) {
		t.Error("map error")
	}
	for i, v := range filtered {
		if v != slice[i]*2 {
			t.Error("map error")
		}
	}
}

func TestFind(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	found, ok := sliceutils.Find(slice, func(item int) bool { return item > 1 })
	if !ok || found != 2 {
		t.Error("find error")
	}
}

func TestSome(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	some := sliceutils.Some(slice, func(item int) bool { return item == 3 })
	if !some {
		t.Error("some error")
	}
}

func TestEvery(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	every := sliceutils.Every(slice, func(item int) bool { return item > 0 })
	if !every {
		t.Error("every error")
	}
}

func TestReduce(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	total := sliceutils.Reduce(slice, func(item, x int) int { return item + x }, 0)
	if total != 15 {
		t.Error("reduce error")
	}
}
