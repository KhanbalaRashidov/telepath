package session

import "testing"

func TestKey(t *testing.T) {
	cases := []struct {
		transport, conv, thread, want string
	}{
		{"telegram", "12345", "", "telegram:12345"},
		{"slack", "C0ABC", "", "slack:C0ABC"},
		{"slack", "C0ABC", "1712.0003", "slack:C0ABC:1712.0003"},
		{"slack", "  ", "", "slack"},
	}
	for _, c := range cases {
		if got := Key(c.transport, c.conv, c.thread); got != c.want {
			t.Errorf("Key(%q,%q,%q) = %q, want %q", c.transport, c.conv, c.thread, got, c.want)
		}
	}
}

func TestMemStoreDistinctKeys(t *testing.T) {
	st := NewMemStore(Defaults{Agent: "claude"})
	a := st.Get(Key("telegram", "1", ""))
	b := st.Get(Key("slack", "1", ""))
	if a == b {
		t.Fatal("telegram and slack conversations with the same id must not collide")
	}
	st.Update(Key("telegram", "1", ""), func(s *State) { s.Agent = "gemini" })
	if st.Get(Key("telegram", "1", "")).Agent != "gemini" {
		t.Fatal("update not persisted")
	}
	if st.Get(Key("slack", "1", "")).Agent != "claude" {
		t.Fatal("slack state should be unaffected by telegram update")
	}
}
