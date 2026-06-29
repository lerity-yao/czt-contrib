package minio

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/zeromicro/go-zero/core/collection"
	"github.com/zeromicro/go-zero/core/logx"

	miniogo "github.com/minio/minio-go/v7"
)

const (
	// decayTime is the EWMA decay period (from finagle defaults).
	decayTime = int64(10 * time.Second)
	// forcePick forces selection of a node if not picked for this duration.
	forcePick = int64(time.Second)
	// affinityTTL is the TTL for write-after-read affinity entries.
	affinityTTL = 5 * time.Second
	// penalty is the load value assigned when no data is available.
	penalty = int64(1 << 31)
)

// node represents a single MinIO endpoint with P2C tracking state.
type node struct {
	client   *miniogo.Client
	endpoint string
	lag      atomic.Uint64 // EWMA latency in nanoseconds
	inflight atomic.Int64  // concurrent in-flight requests
	lastPick atomic.Int64  // last picked UnixNano
}

// load calculates the current load score for the node.
func (n *node) load() int64 {
	// plus one to avoid multiplying by zero
	lag := int64(math.Sqrt(float64(n.lag.Load() + 1)))
	load := lag * (n.inflight.Load() + 1)
	return load
}

// p2cBalancer implements Power-of-Two-Choices load balancing with write-after-read affinity.
type p2cBalancer struct {
	nodes    []*node
	affinity *collection.Cache // "bucket/key" → int (nodeIdx)
	mu       sync.Mutex
	r        *rand.Rand
}

// newP2CBalancer creates a new P2C balancer with the given nodes.
// ak is used as unique identifier for cache metrics isolation.
func newP2CBalancer(nodes []*node, ak string) (*p2cBalancer, error) {
	cache, err := collection.NewCache(affinityTTL, collection.WithName("minio:affinity:"+ak))
	if err != nil {
		return nil, fmt.Errorf("minio: failed to create affinity cache: %w", err)
	}

	return &p2cBalancer{
		nodes:    nodes,
		affinity: cache,
		r:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}, nil
}

// Pick selects a node using P2C algorithm and returns a done callback for EWMA update.
func (b *p2cBalancer) Pick() (*node, func(duration time.Duration)) {
	b.mu.Lock()
	defer b.mu.Unlock()

	var chosen *node
	switch len(b.nodes) {
	case 0:
		return nil, func(time.Duration) {}
	case 1:
		chosen = b.nodes[0]
		chosen.lastPick.Store(time.Now().UnixNano())
	case 2:
		chosen = b.choose(b.nodes[0], b.nodes[1])
	default:
		a := b.r.Intn(len(b.nodes))
		bi := b.r.Intn(len(b.nodes) - 1)
		if bi >= a {
			bi++
		}
		chosen = b.choose(b.nodes[a], b.nodes[bi])
	}

	chosen.inflight.Add(1)
	return chosen, b.buildDone(chosen)
}

// PickWithAffinity checks affinity cache first, falls back to P2C if miss or expired.
func (b *p2cBalancer) PickWithAffinity(bucket, key string) (*node, func(duration time.Duration)) {
	affinityKey := bucket + "/" + key
	if val, ok := b.affinity.Get(affinityKey); ok {
		idx := val.(int)
		if idx >= 0 && idx < len(b.nodes) {
			metricClientAffinityHitTotal.Inc(bucket)
			n := b.nodes[idx]
			n.inflight.Add(1)
			n.lastPick.Store(time.Now().UnixNano())
			return n, b.buildDone(n)
		}
		// invalid index, remove stale entry
		b.affinity.Del(affinityKey)
	}
	// no affinity or invalid index, fall back to P2C
	metricClientAffinityMissTotal.Inc(bucket)
	return b.Pick()
}

// SetAffinity records that a write to bucket/key was served by node n.
func (b *p2cBalancer) SetAffinity(bucket, key string, n *node) {
	affinityKey := bucket + "/" + key
	for i, nd := range b.nodes {
		if nd == n {
			b.affinity.Set(affinityKey, i)
			return
		}
	}
}

// choose picks the less-loaded node between c1 and c2,
// with a forced pick if c2 has been idle too long.
func (b *p2cBalancer) choose(c1, c2 *node) *node {
	start := time.Now().UnixNano()
	if c1.load() > c2.load() {
		c1, c2 = c2, c1
	}

	// if the higher-load node hasn't been picked recently, force-pick it
	pick := c2.lastPick.Load()
	if start-pick > forcePick && c2.lastPick.CompareAndSwap(pick, start) {
		return c2
	}

	c1.lastPick.Store(start)
	return c1
}

// buildDone creates the done callback that decrements inflight and updates EWMA lag.
func (b *p2cBalancer) buildDone(n *node) func(duration time.Duration) {
	return func(duration time.Duration) {
		n.inflight.Add(-1)

		td := duration.Nanoseconds()
		if td < 0 {
			td = 0
		}

		lastLag := n.lag.Load()
		w := math.Exp(float64(-td) / float64(decayTime))
		if lastLag == 0 {
			w = 0
		}
		lag := uint64(float64(lastLag)*w + float64(td)*(1-w))
		n.lag.Store(lag)
	}
}

// nodeIndex returns the index of a node in the balancer's node slice.
func (b *p2cBalancer) nodeIndex(n *node) int {
	for i, nd := range b.nodes {
		if nd == n {
			return i
		}
	}
	return -1
}

// shuffleExcept returns a shuffled slice of node indices, excluding the given index.
func (b *p2cBalancer) shuffleExcept(excludeIdx int) []int {
	b.mu.Lock()
	defer b.mu.Unlock()

	indices := make([]int, 0, len(b.nodes)-1)
	for i := range b.nodes {
		if i != excludeIdx {
			indices = append(indices, i)
		}
	}
	b.r.Shuffle(len(indices), func(i, j int) {
		indices[i], indices[j] = indices[j], indices[i]
	})
	return indices
}

// --- Execute helpers ---

// execute runs fn with P2C selection and network-error failover.
func (c *CommonClient) execute(fn func(*miniogo.Client) error) error {
	b := c.balancer

	// First attempt: P2C picks the optimal node
	n, done := b.Pick()
	if n == nil {
		return fmt.Errorf("minio: no available endpoints")
	}

	start := time.Now()
	err := fn(n.client)
	done(time.Since(start))

	if err == nil {
		return nil
	}
	if !isNetworkError(err) {
		return err
	}

	firstIdx := b.nodeIndex(n)
	metricClientFailoverTotal.Inc(n.endpoint)
	logx.Errorf("minio: endpoint %s failed, trying next: %v", n.endpoint, err)

	// Failover: shuffle remaining nodes and iterate deterministically
	remaining := b.shuffleExcept(firstIdx)
	for _, idx := range remaining {
		nd := b.nodes[idx]
		start = time.Now()
		err = fn(nd.client)
		elapsed := time.Since(start)
		nd.inflight.Add(1)
		b.buildDone(nd)(elapsed)

		if err == nil {
			return nil
		}
		if !isNetworkError(err) {
			return err
		}
		logx.Errorf("minio: endpoint %s failed, trying next: %v", nd.endpoint, err)
	}

	return fmt.Errorf("minio: all %d endpoints failed", len(b.nodes))
}

// executeWith runs fn with P2C selection, failover, and a return value.
func executeWith[T any](c *CommonClient, fn func(*miniogo.Client) (T, error)) (T, error) {
	b := c.balancer
	var zero T

	// First attempt: P2C picks the optimal node
	n, done := b.Pick()
	if n == nil {
		return zero, fmt.Errorf("minio: no available endpoints")
	}

	start := time.Now()
	result, err := fn(n.client)
	done(time.Since(start))

	if err == nil {
		return result, nil
	}
	if !isNetworkError(err) {
		return zero, err
	}

	firstIdx := b.nodeIndex(n)
	metricClientFailoverTotal.Inc(n.endpoint)
	logx.Errorf("minio: endpoint %s failed, trying next: %v", n.endpoint, err)

	// Failover: shuffle remaining nodes and iterate deterministically
	remaining := b.shuffleExcept(firstIdx)
	for _, idx := range remaining {
		nd := b.nodes[idx]
		start = time.Now()
		result, err = fn(nd.client)
		elapsed := time.Since(start)
		nd.inflight.Add(1)
		b.buildDone(nd)(elapsed)

		if err == nil {
			return result, nil
		}
		if !isNetworkError(err) {
			return zero, err
		}
		logx.Errorf("minio: endpoint %s failed, trying next: %v", nd.endpoint, err)
	}

	return zero, fmt.Errorf("minio: all %d endpoints failed", len(b.nodes))
}

// executeWriteWith runs fn with P2C, failover, return value, and sets affinity on success.
func executeWriteWith[T any](c *CommonClient, bucket, key string, fn func(*miniogo.Client) (T, error)) (T, error) {
	b := c.balancer
	var zero T

	// First attempt: P2C picks the optimal node
	n, done := b.Pick()
	if n == nil {
		return zero, fmt.Errorf("minio: no available endpoints")
	}

	start := time.Now()
	result, err := fn(n.client)
	done(time.Since(start))

	if err == nil {
		b.SetAffinity(bucket, key, n)
		return result, nil
	}
	if !isNetworkError(err) {
		return zero, err
	}

	firstIdx := b.nodeIndex(n)
	metricClientFailoverTotal.Inc(n.endpoint)
	logx.Errorf("minio: endpoint %s failed, trying next: %v", n.endpoint, err)

	// Failover: shuffle remaining nodes and iterate deterministically
	remaining := b.shuffleExcept(firstIdx)
	for _, idx := range remaining {
		nd := b.nodes[idx]
		start = time.Now()
		result, err = fn(nd.client)
		elapsed := time.Since(start)
		nd.inflight.Add(1)
		b.buildDone(nd)(elapsed)

		if err == nil {
			b.SetAffinity(bucket, key, nd)
			return result, nil
		}
		if !isNetworkError(err) {
			return zero, err
		}
		logx.Errorf("minio: endpoint %s failed, trying next: %v", nd.endpoint, err)
	}

	return zero, fmt.Errorf("minio: all %d endpoints failed", len(b.nodes))
}

// executeWithAffinityWith runs fn with affinity, failover, and a return value.
func executeWithAffinityWith[T any](c *CommonClient, bucket, key string, fn func(*miniogo.Client) (T, error)) (T, error) {
	b := c.balancer
	n, done := b.PickWithAffinity(bucket, key)
	var zero T
	if n == nil {
		return zero, fmt.Errorf("minio: no available endpoints")
	}

	start := time.Now()
	result, err := fn(n.client)
	done(time.Since(start))

	if err == nil {
		return result, nil
	}
	if !isNetworkError(err) {
		return zero, err
	}

	// Penalize failed node so P2C avoids it during fallback
	n.inflight.Add(penalty)
	defer n.inflight.Add(-penalty)

	metricClientFailoverTotal.Inc(n.endpoint)
	logx.Errorf("minio: affinity endpoint %s failed, falling back: %v", n.endpoint, err)
	return executeWith(c, fn)
}

// pickNode selects a node via P2C for streaming operations (no failover).
func (c *CommonClient) pickNode() *node {
	n, done := c.balancer.Pick()
	if n == nil {
		return nil
	}
	// For streaming ops, we record zero duration since actual duration is unknown.
	done(0)
	return n
}

// pickNodeWithAffinity selects a node with affinity for streaming read operations.
func (c *CommonClient) pickNodeWithAffinity(bucket, key string) *node {
	n, done := c.balancer.PickWithAffinity(bucket, key)
	if n == nil {
		return nil
	}
	done(0)
	return n
}

// --- Network error detection ---

// isNetworkError determines if an error is a network-level error
// that warrants retrying on another endpoint.
func isNetworkError(err error) bool {
	if err == nil {
		return false
	}

	// net.Error covers net.OpError, net.DNSError, and other network errors.
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	// Fallback: check common network error messages for wrapped errors
	// that may not implement net.Error.
	msg := err.Error()
	networkKeywords := []string{
		"connection refused",
		"connection reset",
		"no such host",
		"i/o timeout",
		"network is unreachable",
		"dial tcp",
	}
	for _, kw := range networkKeywords {
		if strings.Contains(strings.ToLower(msg), kw) {
			return true
		}
	}

	return false
}
