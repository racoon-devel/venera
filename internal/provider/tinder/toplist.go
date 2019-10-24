package tinder

import (
	"sort"
	"sync"
)

type ListItem struct {
	ID     uint
	Rating int
}

type topList struct {
	items []ListItem
	mu    sync.Mutex
	size  uint
}

func newTopList(size uint) *topList {
	return &topList{items: make([]ListItem, 0), size: size}
}

func loadTopList(size uint, top []ListItem) *topList {
	list := newTopList(size)
	count := len(top)
	if uint(count) > size {
		count = int(size)
	}

	for i := 0; i < count; i++ {
		list.Push(top[i].ID, top[i].Rating)
	}

	return list
}

func (self *topList) Push(personRecordID uint, rating int) {
	self.mu.Lock()
	defer self.mu.Unlock()

	comparer := func(i, j int) bool {
		return self.items[i].Rating > self.items[j].Rating
	}

	item := ListItem{ID: personRecordID, Rating: rating}

	if uint(len(self.items)) < self.size {
		self.items = append(self.items, item)
		sort.SliceStable(self.items, comparer)
		return
	}

	if item.Rating < self.items[self.size-1].Rating {
		return
	}

	min := 0
	for i, person := range self.items {
		if person.Rating < self.items[min].Rating {
			min = i
		}
	}

	self.items = self.items[:min+copy(self.items[min:], self.items[min+1:])]
	self.items = append(self.items, item)
	sort.SliceStable(self.items, comparer)
}

func (self *topList) Get() []ListItem {
	self.mu.Lock()
	defer self.mu.Unlock()

	result := make([]ListItem, len(self.items))
	copy(result, self.items)
	return result
}

func (self *topList) Clear() {
	self.mu.Lock()
	defer self.mu.Unlock()

	self.items = make([]ListItem, 0)
}
