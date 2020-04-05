package memory

import (
	"container/list"
	"sync"
	"time"
	"web-server/utils/session"
)

var provider = &Provider{list: list.New()}

type SessionStore struct {
	sid          string
	timeAccessed time.Time
	value        map[interface{}]interface{}
}

// Set sets key value pair
func (st *SessionStore) Set(key, value interface{}) {
	st.value[key] = value
	provider.SessionUpdate(st.sid)
	return
}

// Get returns value property corresponding to key
func (st *SessionStore) Get(key interface{}) interface{} {
	provider.SessionUpdate(st.sid)
	if v, ok := st.value[key]; ok {
		return v
	}
	return nil
}

// Delete removes a key, value pair
func (st *SessionStore) Delete(key interface{}) {
	delete(st.value, key)
	provider.SessionUpdate(st.sid)
	return
}

// SessionId returns sessionid
func (st *SessionStore) SessionId() string {
	return st.sid
}

type Provider struct {
	lock     sync.Mutex
	sessions map[string]*list.Element
	list     *list.List
}

func (provider *Provider) SessionInit(sid string) session.Session {
	provider.lock.Lock()
	defer provider.lock.Unlock()
	v := make(map[interface{}]interface{}, 0)
	newsess := &SessionStore{sid: sid, timeAccessed: time.Now(), value: v}
	element := provider.list.PushBack(newsess)
	provider.sessions[sid] = element
	return newsess
}

func (provider *Provider) SessionRead(sid string) session.Session {
	if element, ok := provider.sessions[sid]; ok {
		return element.Value.(*SessionStore)
	}
	return nil
}

func (provider *Provider) SessionDestroy(sid string) {
	if element, ok := provider.sessions[sid]; ok {
		delete(provider.sessions, sid)
		provider.list.Remove(element)
	}
	return
}

func (provider *Provider) SessionGC(maxlifetime int) {
	provider.lock.Lock()
	defer provider.lock.Unlock()

	for {
		element := provider.list.Back()
		if element == nil {
			break
		}

		if (int(element.Value.(*SessionStore).timeAccessed.Unix()) + maxlifetime) < int(time.Now().Unix()) {
			provider.list.Remove(element)
			delete(provider.sessions, element.Value.(*SessionStore).sid)
		} else {
			break
		}
	}
}

func (provider *Provider) SessionUpdate(sid string) {
	provider.lock.Lock()
	defer provider.lock.Unlock()
	if element, ok := provider.sessions[sid]; ok {
		element.Value.(*SessionStore).timeAccessed = time.Now()
		provider.list.MoveToFront(element)
	}
	return
}

func init() {
	provider.sessions = make(map[string]*list.Element, 0)
	session.Register("memory", provider)
}
