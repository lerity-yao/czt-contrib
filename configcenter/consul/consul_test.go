package consul

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type kvPair struct {
	CreateIndex uint64 `json:"CreateIndex"`
	ModifyIndex uint64 `json:"ModifyIndex"`
	LockIndex   uint64 `json:"LockIndex"`
	Key         string `json:"Key"`
	Flags       uint64 `json:"Flags"`
	Value       string `json:"Value"`
	Session     string `json:"Session"`
}

type mockConsulServer struct {
	*httptest.Server
	mu        sync.Mutex
	index     uint64
	value     []byte
	key       string
	changed   chan struct{}
	requestCount int32
}

func newMockConsulServer(t *testing.T, key string, initial []byte) *mockConsulServer {
	m := &mockConsulServer{
		index:   100,
		value:   initial,
		key:     key,
		changed: make(chan struct{}),
	}
	m.Server = httptest.NewServer(http.HandlerFunc(m.handleKV))
	t.Cleanup(m.Close)
	return m
}

func (m *mockConsulServer) handleKV(w http.ResponseWriter, r *http.Request) {
	atomic.AddInt32(&m.requestCount, 1)

	waitIndexStr := r.URL.Query().Get("index")
	waitIndex, _ := strconv.ParseUint(waitIndexStr, 10, 64)

	m.mu.Lock()
	currentIndex := m.index
	currentValue := m.value
	changed := m.changed
	m.mu.Unlock()

	if waitIndex > 0 && waitIndex == currentIndex {
		select {
		case <-changed:
			m.mu.Lock()
			currentIndex = m.index
			currentValue = m.value
			m.mu.Unlock()
		case <-time.After(200 * time.Millisecond):
		}
	}

	w.Header().Set("X-Consul-Index", strconv.FormatUint(currentIndex, 10))
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode([]kvPair{{
		Key:         m.key,
		Value:       base64.StdEncoding.EncodeToString(currentValue),
		CreateIndex: currentIndex,
		ModifyIndex: currentIndex,
	}})
}

func (m *mockConsulServer) updateValue(value []byte) {
	m.mu.Lock()
	m.index++
	m.value = value
	close(m.changed)
	m.changed = make(chan struct{})
	m.mu.Unlock()
}

func TestNewConsulSubscriber_Success(t *testing.T) {
	server := newMockConsulServer(t, "test/key", []byte("name: tom"))

	sub, err := NewConsulSubscriber(ConsulConf{
		Host:   server.URL,
		Scheme: "http",
		Key:    "test/key",
		Type:   "yaml",
	})
	if err != nil {
		t.Fatalf("NewConsulSubscriber error: %v", err)
	}
	defer sub.Stop()

	if sub.Path != "test/key" {
		t.Errorf("Path = %q, want test/key", sub.Path)
	}
	if sub.Type != "yaml" {
		t.Errorf("Type = %q, want yaml", sub.Type)
	}
}

func TestMustNewConsulSubscriber(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("MustNewConsulSubscriber panicked: %v", r)
		}
	}()

	server := newMockConsulServer(t, "test/key", []byte("name: tom"))
	sub := MustNewConsulSubscriber(ConsulConf{
		Host:   server.URL,
		Scheme: "http",
		Key:    "test/key",
		Type:   "yaml",
	})
	defer sub.Stop()

	if sub == nil {
		t.Fatal("MustNewConsulSubscriber returned nil")
	}
}

func TestValue_WithYAML(t *testing.T) {
	server := newMockConsulServer(t, "test/key", []byte("name: tom\nage: 18"))

	sub, err := NewConsulSubscriber(ConsulConf{
		Host:   server.URL,
		Scheme: "http",
		Key:    "test/key",
		Type:   "yaml",
	})
	if err != nil {
		t.Fatalf("NewConsulSubscriber error: %v", err)
	}
	defer sub.Stop()

	v, err := sub.Value()
	if err != nil {
		t.Fatalf("Value() error: %v", err)
	}
	if v == "" {
		t.Fatal("Value() returned empty string")
	}
	if !contains(v, "tom") {
		t.Errorf("Value() = %q, want containing tom", v)
	}
}

func TestValue_WithJSON(t *testing.T) {
	server := newMockConsulServer(t, "test/key", []byte(`{"name":"tom"}`))

	sub, err := NewConsulSubscriber(ConsulConf{
		Host:   server.URL,
		Scheme: "http",
		Key:    "test/key",
		Type:   "json",
	})
	if err != nil {
		t.Fatalf("NewConsulSubscriber error: %v", err)
	}
	defer sub.Stop()

	v, err := sub.Value()
	if err != nil {
		t.Fatalf("Value() error: %v", err)
	}
	if !contains(v, "tom") {
		t.Errorf("Value() = %q, want containing tom", v)
	}
}

func TestValue_KeyNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Consul-Index", "100")
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	sub, err := NewConsulSubscriber(ConsulConf{
		Host:   server.URL,
		Scheme: "http",
		Key:    "missing/key",
		Type:   "yaml",
	})
	if err != nil {
		t.Fatalf("NewConsulSubscriber error: %v", err)
	}
	defer sub.Stop()

	v, err := sub.Value()
	if err != nil {
		t.Fatalf("Value() error: %v", err)
	}
	if v != "" {
		t.Errorf("Value() = %q, want empty string", v)
	}
}

func TestAddListener(t *testing.T) {
	server := newMockConsulServer(t, "test/key", []byte("name: tom"))

	sub, err := NewConsulSubscriber(ConsulConf{
		Host:   server.URL,
		Scheme: "http",
		Key:    "test/key",
		Type:   "yaml",
	})
	if err != nil {
		t.Fatalf("NewConsulSubscriber error: %v", err)
	}
	defer sub.Stop()

	called := make(chan struct{})
	err = sub.AddListener(func() {
		close(called)
	})
	if err != nil {
		t.Fatalf("AddListener error: %v", err)
	}

	// Trigger notification manually through the internal method.
	sub.notifyListeners()

	select {
	case <-called:
		// ok
	case <-time.After(time.Second):
		t.Fatal("listener was not called")
	}
}

func TestWatch_TriggersListener(t *testing.T) {
	server := newMockConsulServer(t, "test/key", []byte("name: tom"))

	sub, err := NewConsulSubscriber(ConsulConf{
		Host:   server.URL,
		Scheme: "http",
		Key:    "test/key",
		Type:   "yaml",
	})
	if err != nil {
		t.Fatalf("NewConsulSubscriber error: %v", err)
	}
	defer sub.Stop()

	called := make(chan struct{}, 1)
	if err := sub.AddListener(func() {
		select {
		case called <- struct{}{}:
		default:
		}
	}); err != nil {
		t.Fatalf("AddListener error: %v", err)
	}

	// Give the watch goroutine time to make the initial request.
	time.Sleep(300 * time.Millisecond)

	// Update the value to bump the index and trigger the listener.
	server.updateValue([]byte("name: jerry"))

	select {
	case <-called:
		// ok
	case <-time.After(2 * time.Second):
		t.Fatal("watch did not trigger listener after value update")
	}
}

func TestStop(t *testing.T) {
	server := newMockConsulServer(t, "test/key", []byte("name: tom"))

	sub, err := NewConsulSubscriber(ConsulConf{
		Host:   server.URL,
		Scheme: "http",
		Key:    "test/key",
		Type:   "yaml",
	})
	if err != nil {
		t.Fatalf("NewConsulSubscriber error: %v", err)
	}

	before := atomic.LoadInt32(&server.requestCount)
	sub.Stop()

	// Wait a bit and ensure no additional long-polling requests are made.
	time.Sleep(500 * time.Millisecond)
	after := atomic.LoadInt32(&server.requestCount)

	// Allow one extra request that may have been in flight.
	if after-before > 2 {
		t.Errorf("watch did not stop, request count grew from %d to %d", before, after)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
