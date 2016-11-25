package auth

import (
	"sync"
	"time"
)

// UserLock maintains a lock to block user from logging in within a short period of time.
type UserLock struct {
	failures map[string]time.Time
	d        time.Duration
	rw       *sync.RWMutex
}

// NewUserLock ...
func NewUserLock(freeze time.Duration) *UserLock {
	return &UserLock{
		make(map[string]time.Time),
		freeze,
		&sync.RWMutex{},
	}
}

// Lock marks a new login failure with the time it happens
func (ul *UserLock) Lock(username string) {
	ul.rw.Lock()
	defer ul.rw.Unlock()
	ul.failures[username] = time.Now()
}

// IsLocked checks whether a login request is happened within a period of time or not
// if it is, the authenticator should ignore the login request and return a failure immediately
func (ul *UserLock) IsLocked(username string) bool {
	ul.rw.RLock()
	defer ul.rw.RUnlock()
	return time.Now().Sub(ul.failures[username]) <= ul.d
}
