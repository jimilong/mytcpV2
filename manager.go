package main

import (
	"sync"
)

const sessionMapNum = 32

type Manager struct {
	sessionMaps [sessionMapNum]sessionMap
	closeOnce   sync.Once
	closeWait   sync.WaitGroup
}

type sessionMap struct {
	sync.RWMutex
	sessions map[uint64]*Session
	closed   bool
}

func NewManager() *Manager {
	manager := &Manager{}
	for i := 0; i < len(manager.sessionMaps); i++ {
		manager.sessionMaps[i].sessions = make(map[uint64]*Session)
	}
	return manager
}

func (manager *Manager) Close() {
	manager.closeOnce.Do(func() {
		for i := 0; i < sessionMapNum; i++ {
			smap := manager.sessionMaps[i]
			smap.Lock()
			smap.closed = true
			for _, session := range smap.sessions {
				session.Close()
			}
			smap.Unlock()
		}
		manager.closeWait.Wait()
	})
}

func (manager *Manager) GetSession(sessionID uint64) *Session {
	smap := manager.sessionMaps[sessionID%sessionMapNum]
	smap.RLock()
	defer smap.RUnlock()

	session, _ := smap.sessions[sessionID]
	return session
}

func (manager *Manager) putSession(session *Session) {
	smap := manager.sessionMaps[session.id%sessionMapNum]
	smap.Lock()
	defer smap.Unlock()

	if smap.closed {
		session.Close()
		return
	}

	smap.sessions[session.id] = session
	manager.closeWait.Add(1)
}

func (manager *Manager) delSession(session *Session) {
	smap := manager.sessionMaps[session.id%sessionMapNum]

	smap.Lock()
	defer smap.Unlock()

	delete(smap.sessions, session.id)
	manager.closeWait.Done()
}
