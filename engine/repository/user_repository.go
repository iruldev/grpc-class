package repository

import (
	"gitlab.com/iruldev/grpc-class/engine/model/entity"
	"sync"
)

type UserRepository interface {
	Save(user *entity.User) error
	Find(username string) (*entity.User, error)
}

type UserRepositoryImpl struct {
	mutex sync.RWMutex
	users map[string]*entity.User
}

func NewUserRepository() UserRepository {
	return &UserRepositoryImpl{
		users: make(map[string]*entity.User),
	}
}

func (r *UserRepositoryImpl) Save(user *entity.User) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.users[user.Username] != nil {
		return ErrAlreadyExists
	}

	r.users[user.Username] = user.Clone()
	return nil
}

func (r *UserRepositoryImpl) Find(username string) (*entity.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	user := r.users[username]
	if user == nil {
		return nil, nil
	}

	return user.Clone(), nil
}
