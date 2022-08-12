package session

import (
	"sync"
	"time"

	"github.com/yzzyx/faktura-pdf/models"
)

// Session describes an active user session
type Session struct {
	User     models.User
	Company  models.Company
	LastSeen time.Time
	ID       string
}

var activeSessions = map[string]*Session{}
var activeSessionMx = &sync.RWMutex{}

// Clear removes a user session from the list of active sessions
func Clear(sessionID string) {
	activeSessionMx.Lock()
	defer activeSessionMx.Unlock()

	delete(activeSessions, sessionID)
}

// Validate checks if the supplied sessionID is active
func Validate(sessionID string) (*Session, bool) {
	activeSessionMx.RLock()
	defer activeSessionMx.RUnlock()

	s, ok := activeSessions[sessionID]
	if !ok {
		return nil, false
	}

	// FIXME - check if LastSeen is before a specified timeout
	s.LastSeen = time.Now()
	return s, ok
}

// New creates a new session and adds it to the list of active sessions
func New(user models.User) (*Session, error) {

	id, err := models.GenerateRandomString(20)
	if err != nil {
		return nil, err
	}

	s := &Session{
		User:     user,
		LastSeen: time.Now(),
		ID:       string(id),
	}

	activeSessionMx.Lock()
	defer activeSessionMx.Unlock()

	activeSessions[s.ID] = s
	return s, nil
}
