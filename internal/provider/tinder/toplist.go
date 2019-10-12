package tinder

import (
	"sort"
	"sync"

	"racoondev.tk/gitea/racoon/venera/internal/types"
)

type topList struct {
	items []types.Person
	mu    sync.Mutex
	size  uint
}

func newTopList(size uint) *topList {
	return &topList{items: make([]types.Person, 0), size: size}
}

func loadTopList(size uint, top []types.Person) *topList {
	list := newTopList(size)
	count := len(top)
	if uint(count) > size {
		count = int(size)
	}

	for i := 0; i < count; i++ {
		list.Push(top[i])
	}

	return list
}

func (self *topList) Push(person types.Person) {
	self.mu.Lock()
	defer self.mu.Unlock()

	comparer := func(i, j int) bool {
		return self.items[i].Rating > self.items[j].Rating
	}

	if uint(len(self.items)) < self.size {
		self.items = append(self.items, person)
		sort.SliceStable(self.items, comparer)
		return
	}

	if person.Rating < self.items[self.size-1].Rating {
		return
	}

	min := 0
	for i, person := range self.items {
		if person.Rating < self.items[min].Rating {
			min = i
		}
	}

	self.items = self.items[:min+copy(self.items[min:], self.items[min+1:])]
	self.items = append(self.items, person)
	sort.SliceStable(self.items, comparer)
}

func (self *topList) Get() []types.Person {
	self.mu.Lock()
	defer self.mu.Unlock()

	result := make([]types.Person, len(self.items))
	copy(result, self.items)
	return result
}
