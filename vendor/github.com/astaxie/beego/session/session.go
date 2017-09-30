// Copyright 2014 beego Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package session provider
//
// Usage:
// import(
//   "github.com/astaxie/beego/session"
// )
//
//	func init() {
//      globalSessions, _ = session.NewManager("memory", `{"cookieName":"gosessionid", "enableSetCookie,omitempty": true, "gclifetime":3600, "maxLifetime": 3600, "secure": false, "cookieLifeTime": 3600, "providerConfig": ""}`)
//		go globalSessions.GC()
//	}
//
// more docs: http://beego.me/docs/module/session.md
package session

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

// Store contains all data for one session process with specific id.
type Store interface {
	Set(key, value interface{}) error     //set session value
	Get(key interface{}) interface{}      //get session value
	Delete(key interface{}) error         //delete session value
	SessionID() string                    //back current sessionID
	SessionRelease(w http.ResponseWriter) // release the resource & save data to provider & return the data
	Flush() error                         //delete all data
}

// Provider contains global session methods and saved SessionStores.
// it can operate a SessionStore by its id.
type Provider interface {
	SessionInit(gclifetime int64, config string) error
	SessionRead(sid string) (Store, error)
	SessionExist(sid string) bool
	SessionRegenerate(oldsid, sid string) (Store, error)
	SessionDestroy(sid string) error
	SessionAll() int //get all active session
	SessionGC()
}

var provides = make(map[string]Provider)

// SLogger a helpful variable to log information about session
var SLogger = NewSessionLog(os.Stderr)

// Register makes a session provide available by the provided name.
// If Register is called twice with the same name or if driver is nil,
// it panics.
func Register(name string, provide Provider) {
	if provide == nil {
		panic("session: Register provide is nil")
	}
	if _, dup := provides[name]; dup {
		panic("session: Register called twice for provider " + name)
	}
	provides[name] = provide
}

type ManagerConfig struct {
	CookieName              string `json:"cookieName"`
	EnableSetCookie         bool   `json:"enableSetCookie,omitempty"`
	Secret                  string `json:"secret"`
	Gclifetime              int64  `json:"gclifetime"`
	Maxlifetime             int64  `json:"maxLifetime"`
	Secure                  bool   `json:"secure"`
	CookieLifeTime          int    `json:"cookieLifeTime"`
	ProviderConfig          string `json:"providerConfig"`
	Domain                  string `json:"domain"`
	SessionIDLength         int64  `json:"sessionIDLength"`
	EnableSidInHttpHeader   bool   `json:"enableSidInHttpHeader"`
	SessionNameInHttpHeader string `json:"sessionNameInHttpHeader"`
	EnableSidInUrlQuery     bool   `json:"enableSidInUrlQuery"`
}

// Manager contains Provider and its configuration.
type Manager struct {
	provider Provider
	config   *ManagerConfig
}

// NewManager Create new Manager with provider name and json config string.
// provider name:
// 1. cookie
// 2. file
// 3. memory
// 4. redis
// 5. mysql
// json config:
// 1. is https  default false
// 2. hashfunc  default sha1
// 3. hashkey default beegosessionkey
// 4. maxage default is none
func NewManager(provideName string, cf *ManagerConfig) (*Manager, error) {
	provider, ok := provides[provideName]
	if !ok {
		return nil, fmt.Errorf("session: unknown provide %q (forgotten import?)", provideName)
	}

	if cf.Maxlifetime == 0 {
		cf.Maxlifetime = cf.Gclifetime
	}

	if cf.EnableSidInHttpHeader {
		if cf.SessionNameInHttpHeader == "" {
			panic(errors.New("SessionNameInHttpHeader is empty"))
		}

		strMimeHeader := textproto.CanonicalMIMEHeaderKey(cf.SessionNameInHttpHeader)
		if cf.SessionNameInHttpHeader != strMimeHeader {
			strErrMsg := "SessionNameInHttpHeader (" + cf.SessionNameInHttpHeader + ") has the wrong format, it should be like this : " + strMimeHeader
			panic(errors.New(strErrMsg))
		}
	}

	err := provider.SessionInit(cf.Maxlifetime, cf.ProviderConfig)
	if err != nil {
		return nil, err
	}

	if cf.SessionIDLength == 0 {
		cf.SessionIDLength = 16
	}

	return &Manager{
		provider,
		cf,
	}, nil
}

// getSid retrieves session identifier from HTTP Request.
// First try to retrieve id by reading from cookie, session cookie name is configurable,
// if not exist, then retrieve id from querying parameters.
//
// error is not nil when there is anything wrong.
// sid is empty when need to generate a new session id
// otherwise return an valid session id.
func (manager *Manager) getSid(r *http.Request) (string, error) {
	cookie, errs := r.Cookie(manager.config.CookieName)
	if errs != nil || cookie.Value == "" {
		var sid string
		if manager.config.EnableSidInUrlQuery {
			errs := r.ParseForm()
			if errs != nil {
				return "", errs
			}

			sid = r.FormValue(manager.config.CookieName)
		}

		// if not found in Cookie / param, then read it from request headers
		if manager.config.EnableSidInHttpHeader && sid == "" {
			sids, isFound := r.Header[manager.config.SessionNameInHttpHeader]
			if isFound && len(sids) != 0 {
				return sids[0], nil
			}
		}

		return sid, nil
	}

	// HTTP Request contains cookie for sessionid info.
	// return url.QueryUnescape(cookie.Value)

	// To be able to share session and cookie with Stratus it needs to use the same algorithm for
	// generating / parsing cookie string
	// https://github.com/expressjs/session/blob/master/index.js#L525
	// https://github.com/tj/node-cookie-signature
	urlDecoded, err := url.QueryUnescape(cookie.Value)
	if err != nil {
		return "", err
	}
	return unsignSessionID(urlDecoded, manager.config.Secret), nil
}

// SessionStart generate or read the session id from http request.
// if session id exists, return SessionStore with this id.
func (manager *Manager) SessionStart(w http.ResponseWriter, r *http.Request) (session Store, err error) {
	sid, errs := manager.getSid(r)
	if errs != nil {
		return nil, errs
	}

	if sid != "" && manager.provider.SessionExist(sid) {
		return manager.provider.SessionRead(sid)
	}

	// Generate a new session
	sid, errs = manager.sessionID()
	if errs != nil {
		return nil, errs
	}

	session, err = manager.provider.SessionRead(sid)
	if err != nil {
		return nil, errs
	}

	// To be able to share session and cookie with Stratus it needs to use the same algorithm for
	// generating / parsing cookie string
	// https://github.com/expressjs/session/blob/master/index.js#L635
	// https://github.com/tj/node-cookie-signature

	cookie := &http.Cookie{
		Name: manager.config.CookieName,
		// Value:    url.QueryEscape(sid),
		Value:    url.QueryEscape(signSessionID(sid, manager.config.Secret)),
		Path:     "/",
		HttpOnly: true,
		Secure:   manager.isSecure(r),
		Domain:   manager.config.Domain,
	}
	if manager.config.CookieLifeTime > 0 {
		cookie.MaxAge = manager.config.CookieLifeTime
		cookie.Expires = time.Now().Add(time.Duration(manager.config.CookieLifeTime) * time.Second)
	}
	if manager.config.EnableSetCookie {
		http.SetCookie(w, cookie)
	}
	r.AddCookie(cookie)

	if manager.config.EnableSidInHttpHeader {
		r.Header.Set(manager.config.SessionNameInHttpHeader, sid)
		w.Header().Set(manager.config.SessionNameInHttpHeader, sid)
	}

	return
}

// SessionDestroy Destroy session by its id in http request cookie.
func (manager *Manager) SessionDestroy(w http.ResponseWriter, r *http.Request) {
	if manager.config.EnableSidInHttpHeader {
		r.Header.Del(manager.config.SessionNameInHttpHeader)
		w.Header().Del(manager.config.SessionNameInHttpHeader)
	}

	cookie, err := r.Cookie(manager.config.CookieName)
	if err != nil || cookie.Value == "" {
		return
	}

	sid, _ := url.QueryUnescape(cookie.Value)
	manager.provider.SessionDestroy(sid)
	if manager.config.EnableSetCookie {
		expiration := time.Now()
		cookie = &http.Cookie{Name: manager.config.CookieName,
			Path:     "/",
			HttpOnly: true,
			Expires:  expiration,
			MaxAge:   -1}

		http.SetCookie(w, cookie)
	}
}

// GetSessionStore Get SessionStore by its id.
func (manager *Manager) GetSessionStore(sid string) (sessions Store, err error) {
	sessions, err = manager.provider.SessionRead(sid)
	return
}

// GC Start session gc process.
// it can do gc in times after gc lifetime.
func (manager *Manager) GC() {
	manager.provider.SessionGC()
	time.AfterFunc(time.Duration(manager.config.Gclifetime)*time.Second, func() { manager.GC() })
}

// SessionRegenerateID Regenerate a session id for this SessionStore who's id is saving in http request.
func (manager *Manager) SessionRegenerateID(w http.ResponseWriter, r *http.Request) (session Store) {
	sid, err := manager.sessionID()
	if err != nil {
		return
	}
	cookie, err := r.Cookie(manager.config.CookieName)
	if err != nil || cookie.Value == "" {
		//delete old cookie
		session, _ = manager.provider.SessionRead(sid)
		cookie = &http.Cookie{Name: manager.config.CookieName,
			Value:    url.QueryEscape(sid),
			Path:     "/",
			HttpOnly: true,
			Secure:   manager.isSecure(r),
			Domain:   manager.config.Domain,
		}
	} else {
		oldsid, _ := url.QueryUnescape(cookie.Value)
		session, _ = manager.provider.SessionRegenerate(oldsid, sid)
		cookie.Value = url.QueryEscape(sid)
		cookie.HttpOnly = true
		cookie.Path = "/"
	}
	if manager.config.CookieLifeTime > 0 {
		cookie.MaxAge = manager.config.CookieLifeTime
		cookie.Expires = time.Now().Add(time.Duration(manager.config.CookieLifeTime) * time.Second)
	}
	if manager.config.EnableSetCookie {
		http.SetCookie(w, cookie)
	}
	r.AddCookie(cookie)

	if manager.config.EnableSidInHttpHeader {
		r.Header.Set(manager.config.SessionNameInHttpHeader, sid)
		w.Header().Set(manager.config.SessionNameInHttpHeader, sid)
	}

	return
}

// GetActiveSession Get all active sessions count number.
func (manager *Manager) GetActiveSession() int {
	return manager.provider.SessionAll()
}

// SetSecure Set cookie with https.
func (manager *Manager) SetSecure(secure bool) {
	manager.config.Secure = secure
}

func (manager *Manager) sessionID() (string, error) {
	b := make([]byte, manager.config.SessionIDLength)
	n, err := rand.Read(b)
	if n != len(b) || err != nil {
		return "", fmt.Errorf("Could not successfully read from the system CSPRNG.")
	}
	return hex.EncodeToString(b), nil
}

// Set cookie with https.
func (manager *Manager) isSecure(req *http.Request) bool {
	if !manager.config.Secure {
		return false
	}
	if req.URL.Scheme != "" {
		return req.URL.Scheme == "https"
	}
	if req.TLS == nil {
		return false
	}
	return true
}

// Log implement the log.Logger
type Log struct {
	*log.Logger
}

// NewSessionLog set io.Writer to create a Logger for session.
func NewSessionLog(out io.Writer) *Log {
	sl := new(Log)
	sl.Logger = log.New(out, "[SESSION]", 1e9)
	return sl
}

// sign session ID for set cookie
// https://github.com/expressjs/session/blob/master/index.js#L634
func signSessionID(sid, secret string) string {
	return "s:" + sid + "." + hashSessionID(sid, secret)
}

// This actually verifies the session ID from cookie
// signed session id from cookie looks like:
// s:FcLdqhSWA29y4JhyiHn4rhJ3bZ_GLuTt.s/INCWYNqDc4ziAx+t+La9+QSuGjnTglvJIhtmMXuUs
// https://github.com/tj/node-cookie-signature/blob/master/index.js#L36
func unsignSessionID(signedSid, secret string) string {
	dotIndex := strings.LastIndex(signedSid, ".")
	sid := signedSid[2:dotIndex]
	hash := signedSid[dotIndex+1:]

	fmt.Println("sid", sid)
	fmt.Println("secret", secret)
	fmt.Println("hash", hash)
	fmt.Println("calculated:", hashSessionID(sid, secret))

	if hashSessionID(sid, secret) == hash {
		fmt.Println("session unsigned")
		return sid
	} else {
		fmt.Println("failed to unsign session")
		return ""
	}
}

// generate hash from session ID
// https://github.com/tj/node-cookie-signature/blob/master/index.js#L16
func hashSessionID(sid, secret string) string {

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(sid))
	hashValue := mac.Sum(nil)

	encoded := base64.StdEncoding.EncodeToString(hashValue)
	re := regexp.MustCompile("=+$")

	result := re.ReplaceAllString(string(encoded), "")

	return result
}
