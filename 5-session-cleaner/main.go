//////////////////////////////////////////////////////////////////////
//
// Given is a SessionManager that stores session information in
// memory. The SessionManager itself is working, however, since we
// keep on adding new sessions to the manager our program will
// eventually run out of memory.
//
// Your task is to implement a session cleaner routine that runs
// concurrently in the background and cleans every session that
// hasn't been updated for more than 5 seconds (of course usually
// session times are much longer).
//
// Note that we expect the session to be removed anytime between 5 and
// 7 seconds after the last update. Also, note that you have to be
// very careful in order to prevent race conditions.
//

package main

import (
	"errors"
	"log"
	"sync"
	"time"
)

// SessionManager keeps track of all sessions from creation, updating
// to destroying.
type SessionManager struct {
	sessions      map[string]Session
	sessionsLock  sync.RWMutex
	expireChannel chan expireMessage
}

// Session stores the session's data
type Session struct {
	Data        map[string]interface{}
	version     uint64
	expireTimer *time.Timer
}

type expireMessage struct {
	sessionID      string
	sessionVersion uint64
}

// NewSessionManager creates a new sessionManager
func NewSessionManager() *SessionManager {
	m := &SessionManager{
		sessions:      make(map[string]Session),
		expireChannel: make(chan expireMessage),
	}

	go m.sessionCleaner()

	return m
}

func (m *SessionManager) sessionCleaner() {
	for {
		msg := <-m.expireChannel

		m.sessionsLock.Lock()
		session := m.sessions[msg.sessionID]
		// Only delete the session if it hasn't been updated concurrently with the expiration message send
		if session.version == msg.sessionVersion {
			delete(m.sessions, msg.sessionID)
		}
		m.sessionsLock.Unlock()
	}
}

func (m *SessionManager) getSessionExpireTimer(sessionID string, sessionVersion uint64) *time.Timer {
	return time.AfterFunc(5*time.Second, func() { m.expireChannel <- expireMessage{sessionID, sessionVersion} })
}

// CreateSession creates a new session and returns the sessionID
func (m *SessionManager) CreateSession() (string, error) {
	sessionID, err := MakeSessionID()
	if err != nil {
		return "", err
	}

	m.sessionsLock.Lock()
	defer m.sessionsLock.Unlock()

	m.sessions[sessionID] = Session{
		Data:        make(map[string]interface{}),
		expireTimer: m.getSessionExpireTimer(sessionID, 0),
	}

	return sessionID, nil
}

// ErrSessionNotFound returned when sessionID not listed in
// SessionManager
var ErrSessionNotFound = errors.New("SessionID does not exists")

// GetSessionData returns data related to session if sessionID is
// found, errors otherwise
func (m *SessionManager) GetSessionData(sessionID string) (map[string]interface{}, error) {
	m.sessionsLock.RLock()
	defer m.sessionsLock.RUnlock()

	session, ok := m.sessions[sessionID]
	if !ok {
		return nil, ErrSessionNotFound
	}
	return session.Data, nil
}

// UpdateSessionData overwrites the old session data with the new one
func (m *SessionManager) UpdateSessionData(sessionID string, data map[string]interface{}) error {
	m.sessionsLock.Lock()
	defer m.sessionsLock.Unlock()

	session, ok := m.sessions[sessionID]
	if !ok {
		return ErrSessionNotFound
	}

	// Hint: you should renew expiry of the session here
	session.expireTimer.Stop()
	newVersion := session.version + 1
	m.sessions[sessionID] = Session{
		Data:        data,
		version:     newVersion,
		expireTimer: m.getSessionExpireTimer(sessionID, newVersion),
	}

	return nil
}

func main() {
	// Create new sessionManager and new session
	m := NewSessionManager()
	sID, err := m.CreateSession()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Created new session with ID", sID)

	// Update session data
	data := make(map[string]interface{})
	data["website"] = "longhoang.de"

	err = m.UpdateSessionData(sID, data)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Update session data, set website to longhoang.de")

	// Retrieve data from manager again
	updatedData, err := m.GetSessionData(sID)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Get session data:", updatedData)
}
