package minio

import (
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	miniogo "github.com/minio/minio-go/v7"
)

// --- node.load() tests ---

func TestNodeLoad_ZeroLag(t *testing.T) {
	n := &node{endpoint: "ep1"}
	l := n.load()
	// lag=0 → sqrt(0+1)=1, inflight=0 → (0+1)=1, load=1*1=1
	if l != 1 {
		t.Fatalf("expected load=1, got %d", l)
	}
}

func TestNodeLoad_WithLagAndInflight(t *testing.T) {
	n := &node{endpoint: "ep1"}
	n.lag.Store(99) // sqrt(100)=10
	n.inflight.Store(4)
	l := n.load()
	// sqrt(99+1)=10, (4+1)=5, 10*5=50
	if l != 50 {
		t.Fatalf("expected load=50, got %d", l)
	}
}

// --- newP2CBalancer tests ---

func TestNewP2CBalancer_Success(t *testing.T) {
	nodes := []*node{{endpoint: "ep1"}, {endpoint: "ep2"}}
	b, err := newP2CBalancer(nodes, "test-ak")
	if err != nil {
		t.Fatal(err)
	}
	if len(b.nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(b.nodes))
	}
	if b.affinity == nil {
		t.Fatal("affinity cache should not be nil")
	}
}

// --- Pick() tests ---

func TestPick_EmptyNodes(t *testing.T) {
	b, _ := newP2CBalancer(nil, "test")
	n, done := b.Pick()
	if n != nil {
		t.Fatal("expected nil node")
	}
	done(0) // should not panic
}

func TestPick_SingleNode(t *testing.T) {
	nodes := []*node{{endpoint: "ep1"}}
	b, _ := newP2CBalancer(nodes, "test")
	n, done := b.Pick()
	if n != nodes[0] {
		t.Fatal("expected the single node")
	}
	if n.inflight.Load() != 1 {
		t.Fatal("expected inflight=1")
	}
	done(10 * time.Millisecond)
	if n.inflight.Load() != 0 {
		t.Fatal("expected inflight=0 after done")
	}
}

func TestPick_TwoNodes(t *testing.T) {
	nodes := []*node{{endpoint: "ep1"}, {endpoint: "ep2"}}
	b, _ := newP2CBalancer(nodes, "test")
	n, done := b.Pick()
	if n == nil {
		t.Fatal("expected non-nil node")
	}
	done(0)
}

func TestPick_MultipleNodes_Distribution(t *testing.T) {
	nodes := make([]*node, 5)
	for i := range nodes {
		nodes[i] = &node{endpoint: fmt.Sprintf("ep%d", i)}
	}
	b, _ := newP2CBalancer(nodes, "test")

	picked := make(map[string]int)
	for i := 0; i < 1000; i++ {
		n, done := b.Pick()
		picked[n.endpoint]++
		done(time.Millisecond)
	}
	// Expect all nodes to be picked at least once
	for _, nd := range nodes {
		if picked[nd.endpoint] == 0 {
			t.Fatalf("node %s was never picked", nd.endpoint)
		}
	}
}

// --- choose() tests ---

func TestChoose_PrefersLowerLoad(t *testing.T) {
	n1 := &node{endpoint: "ep1"}
	n2 := &node{endpoint: "ep2"}
	// Make n1 have higher load
	n1.lag.Store(10000)
	n1.inflight.Store(10)
	// Make n2 have lower load
	n2.lag.Store(0)
	n2.inflight.Store(0)
	// Set recent lastPick on both to avoid forcePick
	now := time.Now().UnixNano()
	n1.lastPick.Store(now)
	n2.lastPick.Store(now)

	b, _ := newP2CBalancer([]*node{n1, n2}, "test")
	chosen := b.choose(n1, n2)
	if chosen != n2 {
		t.Fatal("expected lower-load node to be chosen")
	}
}

func TestChoose_ForcePick_IdleNode(t *testing.T) {
	n1 := &node{endpoint: "ep1"}
	n2 := &node{endpoint: "ep2"}
	// Both have same load
	n1.lag.Store(0)
	n2.lag.Store(0)
	// n2 hasn't been picked for a long time
	n2.lastPick.Store(time.Now().Add(-2 * time.Second).UnixNano())
	n1.lastPick.Store(time.Now().UnixNano())

	b, _ := newP2CBalancer([]*node{n1, n2}, "test")
	chosen := b.choose(n1, n2)
	// n2 should be force-picked since it's been idle > forcePick
	if chosen != n2 {
		t.Fatal("expected idle node to be force-picked")
	}
}

// --- buildDone() EWMA tests ---

func TestBuildDone_UpdatesLag(t *testing.T) {
	nodes := []*node{{endpoint: "ep1"}}
	b, _ := newP2CBalancer(nodes, "test")
	n := nodes[0]
	n.inflight.Store(1)

	done := b.buildDone(n)
	done(100 * time.Millisecond)

	if n.inflight.Load() != 0 {
		t.Fatal("expected inflight decremented")
	}
	lag := n.lag.Load()
	if lag == 0 {
		t.Fatal("expected lag to be updated")
	}
}

func TestBuildDone_NegativeDuration(t *testing.T) {
	nodes := []*node{{endpoint: "ep1"}}
	b, _ := newP2CBalancer(nodes, "test")
	n := nodes[0]
	n.inflight.Store(1)

	done := b.buildDone(n)
	done(-time.Millisecond) // negative should be treated as 0
	// Should not panic and lag should still be updated
}

// --- PickWithAffinity tests ---

func TestPickWithAffinity_Hit(t *testing.T) {
	nodes := []*node{{endpoint: "ep1"}, {endpoint: "ep2"}}
	b, _ := newP2CBalancer(nodes, "test")

	// Set affinity for bucket/key → node 1
	b.affinity.Set("mybucket/mykey", 1)

	n, done := b.PickWithAffinity("mybucket", "mykey")
	done(0)
	if n != nodes[1] {
		t.Fatal("expected affinity hit to return node 1")
	}
}

func TestPickWithAffinity_Miss(t *testing.T) {
	nodes := []*node{{endpoint: "ep1"}, {endpoint: "ep2"}}
	b, _ := newP2CBalancer(nodes, "test")

	n, done := b.PickWithAffinity("mybucket", "nokey")
	done(0)
	if n == nil {
		t.Fatal("expected non-nil node on affinity miss")
	}
}

func TestPickWithAffinity_InvalidIndex(t *testing.T) {
	nodes := []*node{{endpoint: "ep1"}}
	b, _ := newP2CBalancer(nodes, "test")

	// Set invalid index
	b.affinity.Set("mybucket/mykey", 999)

	n, done := b.PickWithAffinity("mybucket", "mykey")
	done(0)
	if n == nil {
		t.Fatal("expected fallback to P2C")
	}
}

// --- SetAffinity tests ---

func TestSetAffinity(t *testing.T) {
	nodes := []*node{{endpoint: "ep1"}, {endpoint: "ep2"}}
	b, _ := newP2CBalancer(nodes, "test")

	b.SetAffinity("bkt", "key1", nodes[0])
	val, ok := b.affinity.Get("bkt/key1")
	if !ok || val.(int) != 0 {
		t.Fatal("expected affinity set to index 0")
	}

	b.SetAffinity("bkt", "key2", nodes[1])
	val, ok = b.affinity.Get("bkt/key2")
	if !ok || val.(int) != 1 {
		t.Fatal("expected affinity set to index 1")
	}
}

func TestSetAffinity_UnknownNode(t *testing.T) {
	nodes := []*node{{endpoint: "ep1"}}
	b, _ := newP2CBalancer(nodes, "test")
	unknown := &node{endpoint: "unknown"}
	b.SetAffinity("bkt", "key", unknown)
	// Should not store anything
	_, ok := b.affinity.Get("bkt/key")
	if ok {
		t.Fatal("expected no affinity for unknown node")
	}
}

// --- nodeIndex tests ---

func TestNodeIndex(t *testing.T) {
	nodes := []*node{{endpoint: "ep1"}, {endpoint: "ep2"}, {endpoint: "ep3"}}
	b, _ := newP2CBalancer(nodes, "test")

	if b.nodeIndex(nodes[0]) != 0 {
		t.Fatal("expected index 0")
	}
	if b.nodeIndex(nodes[2]) != 2 {
		t.Fatal("expected index 2")
	}
	if b.nodeIndex(&node{endpoint: "unknown"}) != -1 {
		t.Fatal("expected -1 for unknown node")
	}
}

// --- shuffleExcept tests ---

func TestShuffleExcept(t *testing.T) {
	nodes := make([]*node, 5)
	for i := range nodes {
		nodes[i] = &node{endpoint: fmt.Sprintf("ep%d", i)}
	}
	b, _ := newP2CBalancer(nodes, "test")

	result := b.shuffleExcept(2)
	if len(result) != 4 {
		t.Fatalf("expected 4 indices, got %d", len(result))
	}
	for _, idx := range result {
		if idx == 2 {
			t.Fatal("excluded index should not be present")
		}
	}
}

func TestShuffleExcept_SingleNode(t *testing.T) {
	nodes := []*node{{endpoint: "ep1"}}
	b, _ := newP2CBalancer(nodes, "test")
	result := b.shuffleExcept(0)
	if len(result) != 0 {
		t.Fatalf("expected 0 indices, got %d", len(result))
	}
}

// --- isNetworkError tests ---

func TestIsNetworkError_Nil(t *testing.T) {
	if isNetworkError(nil) {
		t.Fatal("nil should not be network error")
	}
}

func TestIsNetworkError_NetError(t *testing.T) {
	err := &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("connection refused")}
	if !isNetworkError(err) {
		t.Fatal("net.OpError should be a network error")
	}
}

func TestIsNetworkError_DNSError(t *testing.T) {
	err := &net.DNSError{Err: "no such host", Name: "example.com"}
	if !isNetworkError(err) {
		t.Fatal("net.DNSError should be a network error")
	}
}

func TestIsNetworkError_TimeoutKeyword(t *testing.T) {
	err := errors.New("i/o timeout while connecting")
	if !isNetworkError(err) {
		t.Fatal("i/o timeout should be network error")
	}
}

func TestIsNetworkError_ConnectionRefused(t *testing.T) {
	err := errors.New("dial tcp 127.0.0.1:9000: connection refused")
	if !isNetworkError(err) {
		t.Fatal("connection refused should be network error")
	}
}

func TestIsNetworkError_ConnectionReset(t *testing.T) {
	err := errors.New("connection reset by peer")
	if !isNetworkError(err) {
		t.Fatal("connection reset should be network error")
	}
}

func TestIsNetworkError_NoSuchHost(t *testing.T) {
	err := errors.New("lookup nonexist: no such host")
	if !isNetworkError(err) {
		t.Fatal("no such host should be network error")
	}
}

func TestIsNetworkError_NetworkUnreachable(t *testing.T) {
	err := errors.New("network is unreachable")
	if !isNetworkError(err) {
		t.Fatal("network unreachable should be network error")
	}
}

func TestIsNetworkError_DialTcp(t *testing.T) {
	err := errors.New("dial tcp: some error")
	if !isNetworkError(err) {
		t.Fatal("dial tcp should be network error")
	}
}

func TestIsNetworkError_BusinessError(t *testing.T) {
	err := errors.New("NoSuchKey: object not found")
	if isNetworkError(err) {
		t.Fatal("business error should not be network error")
	}
}

func TestIsNetworkError_MinioErrorResponse(t *testing.T) {
	err := miniogo.ErrorResponse{Code: "NoSuchKey", Message: "not found", StatusCode: 404}
	if isNetworkError(err) {
		t.Fatal("minio ErrorResponse should not be network error")
	}
}

// --- execute / executeWith / executeWriteWith / executeWithAffinityWith tests ---

func newTestCommonClient(nodes []*node) *CommonClient {
	b, _ := newP2CBalancer(nodes, "test")
	return &CommonClient{
		conf:     Conf{Endpoints: []string{"ep1"}},
		balancer: b,
	}
}

func TestExecute_Success(t *testing.T) {
	nodes := []*node{{endpoint: "ep1", client: nil}, {endpoint: "ep2", client: nil}}
	cc := newTestCommonClient(nodes)
	called := false
	err := cc.execute(func(client *miniogo.Client) error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("fn was not called")
	}
}

func TestExecute_BusinessError(t *testing.T) {
	nodes := []*node{{endpoint: "ep1"}, {endpoint: "ep2"}}
	cc := newTestCommonClient(nodes)
	bizErr := errors.New("NoSuchKey")
	err := cc.execute(func(client *miniogo.Client) error {
		return bizErr
	})
	if err != bizErr {
		t.Fatalf("expected business error returned, got: %v", err)
	}
}

func TestExecute_NetworkError_Failover(t *testing.T) {
	nodes := []*node{{endpoint: "ep1"}, {endpoint: "ep2"}, {endpoint: "ep3"}}
	cc := newTestCommonClient(nodes)
	callCount := 0
	err := cc.execute(func(client *miniogo.Client) error {
		callCount++
		if callCount <= 2 {
			return &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("connection refused")}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected success after failover, got: %v", err)
	}
	if callCount < 2 {
		t.Fatalf("expected at least 2 calls, got %d", callCount)
	}
}

func TestExecute_AllFail(t *testing.T) {
	nodes := []*node{{endpoint: "ep1"}, {endpoint: "ep2"}}
	cc := newTestCommonClient(nodes)
	err := cc.execute(func(client *miniogo.Client) error {
		return &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("connection refused")}
	})
	if err == nil {
		t.Fatal("expected error when all nodes fail")
	}
}

func TestExecute_NoNodes(t *testing.T) {
	cc := newTestCommonClient(nil)
	err := cc.execute(func(client *miniogo.Client) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected error for no endpoints")
	}
}

func TestExecuteWith_Success(t *testing.T) {
	nodes := []*node{{endpoint: "ep1"}, {endpoint: "ep2"}}
	cc := newTestCommonClient(nodes)
	result, err := executeWith(cc, func(client *miniogo.Client) (string, error) {
		return "hello", nil
	})
	if err != nil || result != "hello" {
		t.Fatalf("expected 'hello', got %v, err: %v", result, err)
	}
}

func TestExecuteWith_Failover(t *testing.T) {
	nodes := []*node{{endpoint: "ep1"}, {endpoint: "ep2"}, {endpoint: "ep3"}}
	cc := newTestCommonClient(nodes)
	callCount := 0
	result, err := executeWith(cc, func(client *miniogo.Client) (int, error) {
		callCount++
		if callCount == 1 {
			return 0, &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("refused")}
		}
		return 42, nil
	})
	if err != nil || result != 42 {
		t.Fatalf("expected 42, got %v, err: %v", result, err)
	}
}

func TestExecuteWith_BusinessError(t *testing.T) {
	nodes := []*node{{endpoint: "ep1"}}
	cc := newTestCommonClient(nodes)
	bizErr := errors.New("access denied")
	_, err := executeWith(cc, func(client *miniogo.Client) (string, error) {
		return "", bizErr
	})
	if err != bizErr {
		t.Fatalf("expected business error, got: %v", err)
	}
}

func TestExecuteWith_NoNodes(t *testing.T) {
	cc := newTestCommonClient(nil)
	_, err := executeWith(cc, func(client *miniogo.Client) (string, error) {
		return "x", nil
	})
	if err == nil {
		t.Fatal("expected error for no endpoints")
	}
}

func TestExecuteWriteWith_SetsAffinity(t *testing.T) {
	nodes := []*node{{endpoint: "ep1"}, {endpoint: "ep2"}}
	cc := newTestCommonClient(nodes)
	result, err := executeWriteWith(cc, "bkt", "key1", func(client *miniogo.Client) (string, error) {
		return "written", nil
	})
	if err != nil || result != "written" {
		t.Fatalf("expected 'written', got %v, err: %v", result, err)
	}
	// Check affinity was set
	val, ok := cc.balancer.affinity.Get("bkt/key1")
	if !ok {
		t.Fatal("expected affinity to be set")
	}
	idx := val.(int)
	if idx < 0 || idx >= len(nodes) {
		t.Fatalf("invalid affinity index: %d", idx)
	}
}

func TestExecuteWriteWith_Failover_SetsAffinity(t *testing.T) {
	nodes := []*node{{endpoint: "ep1"}, {endpoint: "ep2"}, {endpoint: "ep3"}}
	cc := newTestCommonClient(nodes)
	callCount := 0
	result, err := executeWriteWith(cc, "bkt", "key2", func(client *miniogo.Client) (int, error) {
		callCount++
		if callCount == 1 {
			return 0, &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("refused")}
		}
		return 99, nil
	})
	if err != nil || result != 99 {
		t.Fatalf("expected 99, got %v, err: %v", result, err)
	}
	// Affinity should be set to the successful node
	_, ok := cc.balancer.affinity.Get("bkt/key2")
	if !ok {
		t.Fatal("expected affinity to be set after failover")
	}
}

func TestExecuteWriteWith_NoNodes(t *testing.T) {
	cc := newTestCommonClient(nil)
	_, err := executeWriteWith(cc, "bkt", "key", func(client *miniogo.Client) (string, error) {
		return "x", nil
	})
	if err == nil {
		t.Fatal("expected error for no endpoints")
	}
}

func TestExecuteWithAffinityWith_AffinityHit(t *testing.T) {
	nodes := []*node{{endpoint: "ep1"}, {endpoint: "ep2"}}
	cc := newTestCommonClient(nodes)
	// Set affinity → node 1
	cc.balancer.affinity.Set("bkt/key1", 1)

	result, err := executeWithAffinityWith(cc, "bkt", "key1", func(client *miniogo.Client) (string, error) {
		return "affinity-hit", nil
	})
	if err != nil || result != "affinity-hit" {
		t.Fatalf("expected 'affinity-hit', got %v, err: %v", result, err)
	}
}

func TestExecuteWithAffinityWith_AffinityFallback(t *testing.T) {
	nodes := []*node{{endpoint: "ep1"}, {endpoint: "ep2"}, {endpoint: "ep3"}}
	cc := newTestCommonClient(nodes)
	// Set affinity → node 0
	cc.balancer.affinity.Set("bkt/key1", 0)

	callCount := 0
	result, err := executeWithAffinityWith(cc, "bkt", "key1", func(client *miniogo.Client) (string, error) {
		callCount++
		if callCount == 1 {
			// First call (affinity node) fails with network error
			return "", &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("refused")}
		}
		return "fallback-ok", nil
	})
	if err != nil || result != "fallback-ok" {
		t.Fatalf("expected 'fallback-ok', got %v, err: %v", result, err)
	}
}

func TestExecuteWithAffinityWith_BusinessError(t *testing.T) {
	nodes := []*node{{endpoint: "ep1"}, {endpoint: "ep2"}}
	cc := newTestCommonClient(nodes)
	bizErr := errors.New("not found")
	_, err := executeWithAffinityWith(cc, "bkt", "key", func(client *miniogo.Client) (string, error) {
		return "", bizErr
	})
	if err != bizErr {
		t.Fatalf("expected business error, got: %v", err)
	}
}

func TestExecuteWithAffinityWith_NoNodes(t *testing.T) {
	cc := newTestCommonClient(nil)
	_, err := executeWithAffinityWith(cc, "bkt", "key", func(client *miniogo.Client) (string, error) {
		return "", nil
	})
	if err == nil {
		t.Fatal("expected error for no endpoints")
	}
}

// --- pickNode / pickNodeWithAffinity tests ---

func TestPickNode(t *testing.T) {
	nodes := []*node{{endpoint: "ep1"}, {endpoint: "ep2"}}
	cc := newTestCommonClient(nodes)
	n := cc.pickNode()
	if n == nil {
		t.Fatal("expected non-nil node")
	}
}

func TestPickNode_Empty(t *testing.T) {
	cc := newTestCommonClient(nil)
	n := cc.pickNode()
	if n != nil {
		t.Fatal("expected nil node for empty balancer")
	}
}

func TestPickNodeWithAffinity(t *testing.T) {
	nodes := []*node{{endpoint: "ep1"}, {endpoint: "ep2"}}
	cc := newTestCommonClient(nodes)
	cc.balancer.affinity.Set("bkt/key", 1)

	n := cc.pickNodeWithAffinity("bkt", "key")
	if n != nodes[1] {
		t.Fatal("expected affinity node")
	}
}

func TestPickNodeWithAffinity_Miss(t *testing.T) {
	nodes := []*node{{endpoint: "ep1"}, {endpoint: "ep2"}}
	cc := newTestCommonClient(nodes)
	n := cc.pickNodeWithAffinity("bkt", "nokey")
	if n == nil {
		t.Fatal("expected fallback node")
	}
}

// --- Additional coverage tests ---

func TestExecuteWith_AllFail(t *testing.T) {
	nodes := []*node{{endpoint: "ep1"}, {endpoint: "ep2"}}
	cc := newTestCommonClient(nodes)
	_, err := executeWith(cc, func(client *miniogo.Client) (string, error) {
		return "", &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("refused")}
	})
	if err == nil {
		t.Fatal("expected error when all nodes fail")
	}
}

func TestExecuteWriteWith_BusinessError(t *testing.T) {
	nodes := []*node{{endpoint: "ep1"}}
	cc := newTestCommonClient(nodes)
	bizErr := errors.New("access denied")
	_, err := executeWriteWith(cc, "bkt", "key", func(client *miniogo.Client) (string, error) {
		return "", bizErr
	})
	if err != bizErr {
		t.Fatalf("expected business error, got: %v", err)
	}
	// Verify affinity was NOT set for failed operation
	_, ok := cc.balancer.affinity.Get("bkt/key")
	if ok {
		t.Fatal("affinity should not be set on business error")
	}
}

func TestExecuteWriteWith_AllFail(t *testing.T) {
	nodes := []*node{{endpoint: "ep1"}, {endpoint: "ep2"}}
	cc := newTestCommonClient(nodes)
	_, err := executeWriteWith(cc, "bkt", "key", func(client *miniogo.Client) (int, error) {
		return 0, &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("refused")}
	})
	if err == nil {
		t.Fatal("expected error when all nodes fail")
	}
}

func TestExecute_BusinessErrorOnFailover(t *testing.T) {
	nodes := []*node{{endpoint: "ep1"}, {endpoint: "ep2"}, {endpoint: "ep3"}}
	cc := newTestCommonClient(nodes)
	callCount := 0
	bizErr := errors.New("permission denied")
	err := cc.execute(func(client *miniogo.Client) error {
		callCount++
		if callCount == 1 {
			return &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("refused")}
		}
		return bizErr
	})
	if err != bizErr {
		t.Fatalf("expected business error on failover, got: %v", err)
	}
}

func TestExecuteWith_BusinessErrorOnFailover(t *testing.T) {
	nodes := []*node{{endpoint: "ep1"}, {endpoint: "ep2"}}
	cc := newTestCommonClient(nodes)
	callCount := 0
	bizErr := errors.New("denied")
	_, err := executeWith(cc, func(client *miniogo.Client) (string, error) {
		callCount++
		if callCount == 1 {
			return "", &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("refused")}
		}
		return "", bizErr
	})
	if err != bizErr {
		t.Fatalf("expected business error, got: %v", err)
	}
}

func TestExecuteWriteWith_BusinessErrorOnFailover(t *testing.T) {
	nodes := []*node{{endpoint: "ep1"}, {endpoint: "ep2"}}
	cc := newTestCommonClient(nodes)
	callCount := 0
	bizErr := errors.New("denied")
	_, err := executeWriteWith(cc, "bkt", "key", func(client *miniogo.Client) (string, error) {
		callCount++
		if callCount == 1 {
			return "", &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("refused")}
		}
		return "", bizErr
	})
	if err != bizErr {
		t.Fatalf("expected business error, got: %v", err)
	}
}
