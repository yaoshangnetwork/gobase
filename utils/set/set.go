package set

import (
	"fmt"
	"strings"
)

type Set[T comparable] struct {
	items map[T]struct{}
}

// NewSet 创建一个新的 Set
func NewSet[T comparable]() *Set[T] {
	return &Set[T]{
		items: make(map[T]struct{}),
	}
}

// Add 添加元素
func (s *Set[T]) Add(item T) bool {
	c := s.Contains(item)
	s.items[item] = struct{}{}
	return !c
}

// Remove 删除元素
func (s *Set[T]) Remove(item T) bool {
	if !s.Contains(item) {
		return false
	}
	delete(s.items, item)
	return !s.Contains(item)
}

// Contains 检查元素是否存在
func (s *Set[T]) Contains(item T) bool {
	_, exists := s.items[item]
	return exists
}

// Size 获取 Set 中的元素数量
func (s *Set[T]) Size() int {
	return len(s.items)
}

// ToSlice 转为 slice 结构
func (s *Set[T]) ToSlice() []T {
	slice := make([]T, 0, len(s.items))
	for item := range s.items {
		slice = append(slice, item)
	}
	return slice
}

// Clear 清空 set
func (s *Set[T]) Clear() {
	s.items = make(map[T]struct{})
}

// 实现 String() 方法
func (s *Set[T]) String() string {
	var itemStrings []string
	for item := range s.items {
		itemStrings = append(itemStrings, fmt.Sprintf("%v", item))
	}
	return "{" + strings.Join(itemStrings, ", ") + "}"
}
