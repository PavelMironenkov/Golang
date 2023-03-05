package items

import (
	"strconv"
	"sync"
	"sync/atomic"
)

type ItemMemoryRepository struct {
	lastID uint64
	data   []Item
	mu *sync.RWMutex
}

func NewMemoryRepo() *ItemMemoryRepository {
	return &ItemMemoryRepository{
		data: make([]Item, 0, 10),
		mu: &sync.RWMutex{},
	}
}

func (repo *ItemMemoryRepository) GetAll() ([]Item, error) {
	return repo.data, nil
}

func (repo *ItemMemoryRepository) GetByID(id string) (Item, error) {
	repo.mu.RLock()
	for _, item := range repo.data {
		if item.PostID == id {
			repo.mu.RUnlock()
			return item, nil
		}
	}
	repo.mu.RUnlock()
	return Item{}, nil
}

func (repo *ItemMemoryRepository) Add(item *Item) (lastID uint64,  err error) {
	atomic.AddUint64(&repo.lastID, 1)
	repo.mu.Lock()
	defer repo.mu.Unlock()
	item.PostID = strconv.Itoa(int(repo.lastID))
	repo.data = append(repo.data, *item)
	return repo.lastID, nil
}

func (repo *ItemMemoryRepository) Update(newItem *Item) (bool, error) {
	repo.mu.Lock()
	for i, item := range repo.data {
		if item.PostID != newItem.PostID {
			continue
		}
		repo.data[i].Comments = newItem.Comments
		repo.data[i].Score = newItem.Score
		repo.data[i].UpvotePercentage = newItem.UpvotePercentage
		repo.data[i].Views = newItem.Views
		repo.data[i].Votes = newItem.Votes
		repo.mu.Unlock()
		return true, nil
	}
	return false, nil
}

func (repo *ItemMemoryRepository) Delete(id string) (bool, error) {
	i := -1
	repo.mu.Lock()
	for idx, item := range repo.data {
		if item.PostID != id {
			continue
		}
		i = idx
	}
	if i < 0 {
		return false, nil
	}
	if i < len(repo.data)-1 {
		copy(repo.data[i:], repo.data[i+1:])
	}
	// repo.data[len(repo.data)-1] = nil // or the zero value of T
	repo.data = repo.data[:len(repo.data)-1]
	repo.mu.Unlock()
	return true, nil
}
