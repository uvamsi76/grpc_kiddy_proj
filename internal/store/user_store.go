// internal/store/user_store.go
package store

import (
	"fmt"
	"sync"
	"time"

	pb "github.com/uvamsi76/grpc_kiddy_proj/gen/user"
)

// UserStore is a thread-safe in-memory store for users
type UserStore struct {
	mu    sync.RWMutex // RWMutex: many readers OR one writer at a time
	users map[string]*pb.User
}

func NewUserStore() *UserStore {
	return &UserStore{
		users: make(map[string]*pb.User),
	}
}

// generateID creates a simple unique ID
func generateID() string {
	return fmt.Sprintf("user_%d", time.Now().UnixNano())
}

func (s *UserStore) Create(user *pb.User) *pb.User {
	s.mu.Lock() // exclusive write lock
	defer s.mu.Unlock()

	user.Id = generateID()
	user.Active = true
	s.users[user.Id] = user
	return user
}

func (s *UserStore) GetByID(id string) (*pb.User, bool) {
	s.mu.RLock() // shared read lock — multiple readers OK
	defer s.mu.RUnlock()

	user, ok := s.users[id]
	return user, ok
}

func (s *UserStore) GetByEmail(email string) (*pb.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, u := range s.users {
		if u.Email == email {
			return u, true
		}
	}
	return nil, false
}

func (s *UserStore) Update(id string, applyFn func(*pb.User)) (*pb.User, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, ok := s.users[id]
	if !ok {
		return nil, false
	}
	applyFn(user) // caller decides what to change
	return user, true
}

func (s *UserStore) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.users[id]
	if !ok {
		return false
	}
	delete(s.users, id)
	return true
}

func (s *UserStore) List(page, pageSize int32) ([]*pb.User, int32) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Collect all users into a slice
	all := make([]*pb.User, 0, len(s.users))
	for _, u := range s.users {
		all = append(all, u)
	}

	total := int32(len(all))

	// Apply pagination
	if pageSize <= 0 {
		pageSize = 10
	}
	if page <= 0 {
		page = 1
	}

	start := (page - 1) * pageSize
	if start >= total {
		return []*pb.User{}, total
	}

	end := start + pageSize
	if end > total {
		end = total
	}

	return all[start:end], total
}
