package session

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// Manager is the interface for session manager
type Manager struct {
	cookieName  string
	lock        sync.Mutex
	provider    Provider
	maxlifetime int
}

// Provider is the interface for providers
type Provider interface {
	SessionInit(sid string) Session
	SessionRead(sid string) Session
	SessionDestroy(sid string)
	SessionGC(maxlifetime int)
}

// Session is the interface for all the sessions
type Session interface {
	Set(key, value interface{})
	Get(key interface{}) interface{}
	Delete(key interface{})
	SessionId() string
}

// NewManager creates a new session manager and returns its pointer reference
func NewManager(providerName, cookieName string, maxlifetime int) (*Manager, error) {
	provider, ok := provides[providerName]
	if !ok {
		return nil, fmt.Errorf("session: unknown provide %q (forgotten import?)", providerName)
	}
	return &Manager{provider: provider, cookieName: cookieName, maxlifetime: maxlifetime}, nil
}

var provides = make(map[string]Provider)

// Register creates a session provider with the provided name
func Register(name string, provider Provider) {
	if provider == nil {
		fmt.Println("session: Register provider is nil")
		return
	}
	if _, dup := provides[name]; dup {
		fmt.Println("session: Register called twice for ", name)
		return
	}
	provides[name] = provider
}

func (manager *Manager) sessionId() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

// SessionStart checks if any session associated with the request exists and creates assigns a new one if not present
func (manager *Manager) SessionStart(w http.ResponseWriter, r *http.Request) Session {
	manager.lock.Lock()
	defer manager.lock.Unlock()
	session, sessionExists := manager.SessionCheck(r)
	if sessionExists {
		return session
	}
	sid := manager.sessionId()
	session = manager.provider.SessionInit(sid)
	cookie := http.Cookie{Name: manager.cookieName, Value: url.QueryEscape(sid), Path: "/", HttpOnly: true, MaxAge: manager.maxlifetime}
	http.SetCookie(w, &cookie)
	return session
}

// SessionCheck checks if any session associated with the request exists and returns the session if found
func (manager *Manager) SessionCheck(r *http.Request) (Session, bool) {
	cookie, err := r.Cookie(manager.cookieName)
	if err != nil || cookie.Value == "" {
		return nil, false
	} else {
		sid, _ := url.QueryUnescape(cookie.Value)
		session := manager.provider.SessionRead(sid)
		if session != nil {
			return session, true
		}
		return session, false
	}
}

// GetCookie returns the cookie set associated with the request
func (manager *Manager) GetCookie(r *http.Request) (*http.Cookie, error) {
	cookie, err := r.Cookie(manager.cookieName)
	if err != nil || cookie.Value == "" {
		return nil, err
	}
	return cookie, nil
}

// SessionDestroy deletes any existing session associated with the request
func (manager *Manager) SessionDestroy(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(manager.cookieName)
	if err != nil || cookie.Value == "" {
		return
	} else {
		manager.lock.Lock()
		defer manager.lock.Unlock()
		manager.provider.SessionDestroy(cookie.Value)
		expiration := time.Now()
		cookie := http.Cookie{Name: manager.cookieName, Path: "/", HttpOnly: true, Expires: expiration, MaxAge: -1}
		http.SetCookie(w, &cookie)
	}
}

// GC deletes sessions after their allowed lifetime
func (manager *Manager) GC() {
	manager.lock.Lock()
	defer manager.lock.Unlock()
	manager.provider.SessionGC(manager.maxlifetime)
	time.AfterFunc(time.Duration(manager.maxlifetime), func() { manager.GC() })
}
