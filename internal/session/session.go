// Package session holds per-conversation state: the chosen agent, permission
// and branch modes, the last session id (for resume), and the currently running
// task. State is keyed by a transport-qualified conversation key so Telegram
// chats and Slack channels/DMs never collide. v1 is in-memory behind the Store
// interface so a persistent store can be added later without touching callers.
package session

import (
	"context"
	"strings"
	"sync"

	"github.com/telepath/telepath/internal/agent"
)

// Key composes a stable store key from the transport name, the conversation id,
// and an optional thread id. Empty components are skipped, so Telegram chats
// (no thread) and Slack channels/threads produce distinct, readable keys:
//
//	Key("telegram", "12345", "")      == "telegram:12345"
//	Key("slack", "C0ABC", "")         == "slack:C0ABC"
//	Key("slack", "C0ABC", "171.000")  == "slack:C0ABC:171.000"
func Key(transport, conversationID, threadID string) string {
	parts := make([]string, 0, 3)
	for _, p := range []string{transport, conversationID, threadID} {
		if strings.TrimSpace(p) != "" {
			parts = append(parts, p)
		}
	}
	return strings.Join(parts, ":")
}

// State is the mutable per-conversation state.
type State struct {
	Key string // the store key (transport:conversation[:thread])

	Agent          string               // selected agent name
	PermissionMode agent.PermissionMode // selected permission mode
	BranchMode     agent.BranchMode     // selected branch mode

	LastSessionID string // last agent session id, for resume

	// Running task fields (set while a task is in flight).
	Running    bool
	TaskBranch string
	BaseBranch string
	Cancel     context.CancelFunc
}

// Store is the per-conversation state store.
type Store interface {
	// Get returns the conversation's state, creating it with defaults if absent.
	Get(key string) *State
	// Update applies fn to the conversation's state under lock.
	Update(key string, fn func(*State))
}

// Defaults are applied to a freshly created conversation state.
type Defaults struct {
	Agent          string
	PermissionMode agent.PermissionMode
	BranchMode     agent.BranchMode
}

// memStore is the default in-memory Store.
type memStore struct {
	mu       sync.Mutex
	defaults Defaults
	byKey    map[string]*State
}

// NewMemStore creates an in-memory store seeded with the given defaults.
func NewMemStore(d Defaults) Store {
	return &memStore{
		defaults: d,
		byKey:    make(map[string]*State),
	}
}

func (m *memStore) Get(key string) *State {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.getLocked(key)
}

func (m *memStore) getLocked(key string) *State {
	s, ok := m.byKey[key]
	if !ok {
		s = &State{
			Key:            key,
			Agent:          m.defaults.Agent,
			PermissionMode: m.defaults.PermissionMode,
			BranchMode:     m.defaults.BranchMode,
		}
		m.byKey[key] = s
	}
	return s
}

func (m *memStore) Update(key string, fn func(*State)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	fn(m.getLocked(key))
}
