package auth

import (
	"fmt"
	"os"
	"time"
	"utils/log"

	"models"
)

// 1.5 seconds
const frozenTime time.Duration = 1500 * time.Millisecond

var lock = NewUserLock(frozenTime)

// Authenticator provides interface to authenticate user credentials.
type Authenticator interface {

	// Authenticate ...
	Authenticate(m models.AuthModel) (*models.User, error)
}

var registry = make(map[string]Authenticator)

// Register add different authenticators to registry map.
func Register(name string, authenticator Authenticator) {
	if _, dup := registry[name]; dup {
		log.Infof("authenticator: %s has been registered", name)
		return
	}
	registry[name] = authenticator
}

// Login authenticates user credentials based on setting.
func Login(m models.AuthModel) (*models.User, error) {

	var authMode = os.Getenv("AUTH_MODE")
	if authMode == "" || m.Principal == "admin" {
		authMode = "db_auth"
	}
	log.Debug("Current AUTH_MODE is ", authMode)

	authenticator, ok := registry[authMode]
	if !ok {
		return nil, fmt.Errorf("Unrecognized auth_mode: %s", authMode)
	}
	if lock.IsLocked(m.Principal) {
		log.Debugf("%s is locked due to login failure, login failed", m.Principal)
		return nil, nil
	}
	user, err := authenticator.Authenticate(m)
	if user == nil && err == nil {
		log.Debugf("Login failed, locking %s, and sleep for %v", m.Principal, frozenTime)
		lock.Lock(m.Principal)
		time.Sleep(frozenTime)
	}
	return user, err
}
