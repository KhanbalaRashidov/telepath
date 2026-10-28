// Package session holds per-chat state: the chosen agent, permission and branch
// modes, the last session id (for /resume), and the currently running task.
// v1 is in-memory behind the Store interface so a persistent store can be added
// later without touching callers.
package session

import (
	"context"
	"sync"

	"github.com/telepath/telepath/internal/agent"
)

// State is the mutable per-chat state.
type State struct {
	ChatID int64

	Agent          string               // selected agent name
	PermissionMode agent.PermissionMode // selected permission mode
	BranchMode     agent.BranchMode     // selected branch mode

	LastSessionID string // last agent session id, for /resume

	// Running task fields (set while a task is in flight).
	Running    bool
	TaskBranch string
	BaseBranch string
	Cancel     context.CancelFunc
}

// BranchMode lives in the agent package to avoid an import cycle for the UI; it
// is re-exported here for convenience of session callers.
// (agent.BranchMode is the canonical type.)

// Store is the per-chat state store.
type Store interface {
	// Get returns the chat's state, creating it with defaults if absent.
	Get(chatID int64) *State
	// Update applies fn to the chat's state under lock.
	Update(chatID int64, fn func(*State))
}

// Defaults are applied to a freshly created chat state.
type Defaults struct {
	Agent          string
	PermissionMode agent.PermissionMode
	BranchMode     agent.BranchMode
}

// memStore is the default in-memory Store.
type memStore struct {
	mu       sync.Mutex
	defaults Defaults
	byChat   map[int64]*State
}

// NewMemStore creates an in-memory store seeded with the given defaults.
func NewMemStore(d Defaults) Store {
	return &memStore{
		defaults: d,
		byChat:   make(map[int64]*State),
	}
}

func (m *memStore) Get(chatID int64) *State {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.getLocked(chatID)
}

func (m *memStore) getLocked(chatID int64) *State {
	s, ok := m.byChat[chatID]
	if !ok {
		s = &State{
			ChatID:         chatID,
			Agent:          m.defaults.Agent,
			PermissionMode: m.defaults.PermissionMode,
			BranchMode:     m.defaults.BranchMode,
		}
		m.byChat[chatID] = s
	}
	return s
}

func (m *memStore) Update(chatID int64, fn func(*State)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	fn(m.getLocked(chatID))
}
