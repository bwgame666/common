package ws

import (
	"sync"
	"sync/atomic"
)

type ISessionManager interface {
	Add(IWebSocketSession)
	Remove(IWebSocketSession)
	Get(int64) IWebSocketSession
	ForEach(func(IWebSocketSession) bool)
	Count() int
	CloseAll()
	SetIDSeed(seed int64)
}

type SessionManager struct {
	sesIdSeed int64
	count     int64
	sessById  sync.Map
}

func NewSessionManager(seed int64) ISessionManager {
	return &SessionManager{sesIdSeed: seed, count: 0}
}

func (s *SessionManager) SetIDSeed(seed int64) {
	atomic.StoreInt64(&s.sesIdSeed, seed)
}

func (s *SessionManager) Add(ses IWebSocketSession) {
	id := atomic.AddInt64(&s.sesIdSeed, 1)

	ses.SetID(id)

	atomic.AddInt64(&s.count, 1)
	s.sessById.Store(id, ses)
}

func (s *SessionManager) Remove(ses IWebSocketSession) {
	s.sessById.Delete(ses.ID())

	atomic.AddInt64(&s.count, -1)
}

func (s *SessionManager) Get(id int64) IWebSocketSession {
	if v, ok := s.sessById.Load(id); ok {
		return v.(IWebSocketSession)
	}

	return nil
}

func (s *SessionManager) ForEach(callback func(IWebSocketSession) bool) {
	s.sessById.Range(func(key, value interface{}) bool {
		return callback(value.(IWebSocketSession))
	})
}

func (s *SessionManager) CloseAll() {
	s.ForEach(func(ses IWebSocketSession) bool {
		ses.Close()
		return true
	})
}

func (s *SessionManager) Count() int {
	v := atomic.LoadInt64(&s.count)
	return int(v)
}
