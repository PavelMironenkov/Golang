package user

import (
	"errors"
	"sync"
	"sync/atomic"
)

var (
	ErrExistUser  = errors.New("already exist")
	ErrNoUser  = errors.New("user not found")
	ErrBadPass = errors.New("invalid password")
)

type UserMemoryRepository struct {
	currFreeID uint64
	data map[string]*User
	mu *sync.RWMutex
}

func NewMemoryRepo() *UserMemoryRepository {
	return &UserMemoryRepository{
		data: make(map[string]*User, 0),
		mu:  &sync.RWMutex{},
		currFreeID: 0,
	}
}

func (repo *UserMemoryRepository) Authorize(login, pass string) (*User, error) {
	repo.mu.Lock()
	u, ok := repo.data[login]
	repo.mu.Unlock()
	if !ok {
		return nil, ErrNoUser
	}
	if u.password != pass {
		return nil, ErrBadPass
	}
	return u, nil
}

func(repo *UserMemoryRepository) Register(login, pass string) (*User, error){
	user := &User{Login: login, password: pass, ID: repo.currFreeID}
	repo.mu.Lock()
	_ , exist := repo.data[login]
	if(!exist){
		repo.data[login] = user
		repo.mu.Unlock()
		atomic.AddUint64(&repo.currFreeID, 1)
	}else{
		repo.mu.Unlock()
		return nil, ErrExistUser
	}
	
	return user, nil
}