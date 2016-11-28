package auth

import (
	"fmt"
	"os"
	"time"
	"utils/log"

	"models"
	"service/auth/dashboard"
	"service/auth/db"
	"service/auth/ldap"
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

func init() {
	Register("db_auth", &db.Auth{})
	Register("ldap_auth", &ldap.Auth{})
	Register("dashboard", &dashboard.Auth{})
}

// Register add different authenticators to registry map.
func Register(name string, authenticator Authenticator) {
	log.Debugf("about to register authenticator with name: %s", name)
	if _, dup := registry[name]; dup {
		log.Warningf("authenticator: %s has been registered", name)
		return
	}
	registry[name] = authenticator
}

// Login authenticates user credentials based on setting.
func Login(m models.AuthModel) (*models.User, error) {

	var authMode = os.Getenv("AUTH_MODE")
	if authMode == "" || m.Principal == "admin" {
		authMode = "dashboard"
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
