package consul

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"
	"google.golang.org/grpc/resolver"
)

// ─────────────────────────────────────────────
// config.go – Validate()
// ─────────────────────────────────────────────

func TestValidate_EmptyHost(t *testing.T) {
	c := Conf{Key: "svc"}
	err := c.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty consul hosts")
}

func TestValidate_EmptyKey(t *testing.T) {
	c := Conf{Host: "localhost:8500"}
	err := c.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty consul key")
}

func TestValidate_DefaultsApplied(t *testing.T) {
	c := Conf{Host: "localhost:8500", Key: "svc"}
	require.NoError(t, c.Validate())
	assert.Equal(t, CheckTypeTTL, c.CheckType)
	assert.Equal(t, 20, c.TTL)
	assert.Equal(t, 3, c.ExpiredTTL)
	assert.Equal(t, 3, c.CheckTimeout)
	assert.Equal(t, "http", c.Scheme)
}

func TestValidate_CheckTypeTTL(t *testing.T) {
	c := Conf{Host: "h:8500", Key: "k", CheckType: CheckTypeTTL}
	require.NoError(t, c.Validate())
}

func TestValidate_CheckTypeGrpc(t *testing.T) {
	c := Conf{Host: "h:8500", Key: "k", CheckType: CheckTypeGrpc}
	require.NoError(t, c.Validate())
}

func TestValidate_CheckTypeHttp_DefaultsApplied(t *testing.T) {
	c := Conf{Host: "h:8500", Key: "k", CheckType: CheckTypeHttp}
	require.NoError(t, c.Validate())
	assert.Equal(t, "http", c.CheckHttp.Scheme)
	assert.Equal(t, "GET", c.CheckHttp.Method)
	assert.Equal(t, healthPath, c.CheckHttp.Path)
	assert.Equal(t, healthPort, c.CheckHttp.Port)
	assert.Equal(t, "0.0.0.0", c.CheckHttp.Host)
}

func TestValidate_CheckTypeHttp_ExistingValues(t *testing.T) {
	c := Conf{
		Host:      "h:8500",
		Key:       "k",
		CheckType: CheckTypeHttp,
		CheckHttp: CheckHttpConf{Scheme: "https", Method: "POST", Path: "/ping", Port: 9090, Host: "1.2.3.4"},
	}
	require.NoError(t, c.Validate())
	assert.Equal(t, "https", c.CheckHttp.Scheme)
	assert.Equal(t, "POST", c.CheckHttp.Method)
	assert.Equal(t, "/ping", c.CheckHttp.Path)
	assert.Equal(t, 9090, c.CheckHttp.Port)
	assert.Equal(t, "1.2.3.4", c.CheckHttp.Host)
}

func TestValidate_UnknownCheckType(t *testing.T) {
	c := Conf{Host: "h:8500", Key: "k", CheckType: "unknown"}
	err := c.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown check type")
}

// ─────────────────────────────────────────────
// target.go – parseURL / consulConfig / String
// ─────────────────────────────────────────────

func mustParseURL(raw string) url.URL {
	u, err := url.Parse(raw)
	if err != nil {
		panic(err)
	}
	return *u
}

func TestParseURL_Happy(t *testing.T) {
	u := mustParseURL("consul://localhost:8500/my-service?tag=v1")
	tgt, err := parseURL(u)
	require.NoError(t, err)
	assert.Equal(t, "localhost:8500", tgt.Addr)
	assert.Equal(t, "my-service", tgt.Service)
	assert.Equal(t, "v1", tgt.Tag)
	assert.Equal(t, "_agent", tgt.Near)         // default
	assert.Equal(t, time.Second, tgt.MaxBackoff) // default
}

func TestParseURL_WithCredentials(t *testing.T) {
	u := mustParseURL("consul://user:pass@localhost:8500/svc")
	tgt, err := parseURL(u)
	require.NoError(t, err)
	assert.Equal(t, "user", tgt.User)
	assert.Equal(t, "pass", tgt.Password)
}

func TestParseURL_BadScheme(t *testing.T) {
	u := mustParseURL("http://localhost:8500/svc")
	_, err := parseURL(u)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Malformed URL")
}

func TestParseURL_MissingHost(t *testing.T) {
	u := url.URL{Scheme: "consul", Path: "/svc"}
	_, err := parseURL(u)
	require.Error(t, err)
}

func TestParseURL_MissingPath(t *testing.T) {
	u := url.URL{Scheme: "consul", Host: "localhost:8500"}
	_, err := parseURL(u)
	require.Error(t, err)
}

func TestParseURL_NearAndMaxBackoffDefaults(t *testing.T) {
	u := mustParseURL("consul://localhost:8500/svc")
	tgt, err := parseURL(u)
	require.NoError(t, err)
	assert.Equal(t, "_agent", tgt.Near)
	assert.Equal(t, time.Second, tgt.MaxBackoff)
}

func TestTarget_String(t *testing.T) {
	tgt := target{Service: "my-svc", Healthy: true, Tag: "prod"}
	s := tgt.String()
	assert.Contains(t, s, "my-svc")
	assert.Contains(t, s, "true")
	assert.Contains(t, s, "prod")
}

func TestTarget_ConsulConfig_NoCredentials(t *testing.T) {
	tgt := target{Addr: "localhost:8500", Timeout: 5 * time.Second}
	cfg := tgt.consulConfig()
	assert.Equal(t, "localhost:8500", cfg.Address)
	assert.Nil(t, cfg.HttpAuth)
	assert.Equal(t, 5*time.Second, cfg.HttpClient.Timeout)
}

func TestTarget_ConsulConfig_WithCredentials(t *testing.T) {
	tgt := target{Addr: "localhost:8500", User: "u", Password: "p", Token: "tok", TLSInsecure: true}
	cfg := tgt.consulConfig()
	require.NotNil(t, cfg.HttpAuth)
	assert.Equal(t, "u", cfg.HttpAuth.Username)
	assert.Equal(t, "p", cfg.HttpAuth.Password)
	assert.Equal(t, "tok", cfg.Token)
	assert.True(t, cfg.TLSConfig.InsecureSkipVerify)
}

// ─────────────────────────────────────────────
// register.go – figureOutListenOn
// ─────────────────────────────────────────────

func TestFigureOutListenOn_NonAllEths(t *testing.T) {
	result := figureOutListenOn("10.0.0.1:8080")
	assert.Equal(t, "10.0.0.1:8080", result)
}

func TestFigureOutListenOn_WithPodIP(t *testing.T) {
	t.Setenv(envPodIP, "192.168.1.100")
	result := figureOutListenOn("0.0.0.0:8080")
	assert.Equal(t, "192.168.1.100:8080", result)
}

func TestFigureOutListenOn_EmptyHost(t *testing.T) {
	// Just the port part, no host
	result := figureOutListenOn(":8080")
	// host is empty string (not allEths), returns as-is
	assert.Contains(t, result, "8080")
}

func TestFigureOutListenOn_AllEthsNoPodIPNoInternalIP(t *testing.T) {
	// Unset POD_IP; netx.InternalIp() may return a real IP or empty.
	// We only test that when POD_IP is set, it is preferred.
	os.Unsetenv(envPodIP)
	result := figureOutListenOn("0.0.0.0:9999")
	// Result is either real IP or original, both are valid.
	assert.Contains(t, result, "9999")
}

// ─────────────────────────────────────────────
// register.go – MonitorState.Close
// ─────────────────────────────────────────────

func TestMonitorState_Close_WithTicker(t *testing.T) {
	ms := &MonitorState{
		Ticker: time.NewTicker(time.Hour),
	}
	ms.Close()
	assert.Nil(t, ms.Ticker)
}

func TestMonitorState_Close_NilTicker(t *testing.T) {
	ms := &MonitorState{}
	// Should not panic
	ms.Close()
	assert.Nil(t, ms.Ticker)
}

func TestMonitorState_Close_Idempotent(t *testing.T) {
	ms := &MonitorState{Ticker: time.NewTicker(time.Hour)}
	ms.Close()
	ms.Close() // second call should not panic
	assert.Nil(t, ms.Ticker)
}

// ─────────────────────────────────────────────
// register.go – clientRegistration / NewService via mocked Consul
// ─────────────────────────────────────────────

// fakeConsulServer returns a minimal httptest server that handles the
// Consul Agent API endpoints used by register.go/NewService.
func fakeConsulServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	// PUT /v1/agent/service/register
	mux.HandleFunc("/v1/agent/service/register", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	// PUT /v1/agent/service/deregister/{id}
	mux.HandleFunc("/v1/agent/service/deregister/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	// PUT /v1/agent/check/register
	mux.HandleFunc("/v1/agent/check/register", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	// PUT /v1/agent/check/update/{id}
	mux.HandleFunc("/v1/agent/check/update/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	// GET /v1/agent/service/{id}
	mux.HandleFunc("/v1/agent/service/", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"ID":      "svc-id",
			"Service": "test-svc",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	// GET /v1/health/service/{name}
	mux.HandleFunc("/v1/health/service/", func(w http.ResponseWriter, r *http.Request) {
		entries := []*api.ServiceEntry{
			{
				Service: &api.AgentService{ID: "test-svc-127.0.0.1-8000", Service: "test-svc"},
				Checks:  api.HealthChecks{{Status: api.HealthPassing}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(entries)
	})

	return httptest.NewServer(mux)
}

func TestNewService_ValidateFails(t *testing.T) {
	// Empty host → Validate returns error
	_, err := NewService("127.0.0.1:8000", Conf{Key: "svc"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty consul hosts")
}

func TestNewService_BadListenOnAddress(t *testing.T) {
	srv := fakeConsulServer(t)
	defer srv.Close()

	host := srv.Listener.Addr().String()
	conf := Conf{Host: host, Key: "svc", CheckType: CheckTypeTTL, TTL: 20, ExpiredTTL: 3, CheckTimeout: 3, Scheme: "http"}
	// Pass a listenOn that can't be split into host:port
	_, err := NewService("not-a-valid-addr", conf)
	require.Error(t, err)
}

func TestNewService_TTL_Success(t *testing.T) {
	srv := fakeConsulServer(t)
	defer srv.Close()

	host := srv.Listener.Addr().String()
	conf := Conf{Host: host, Key: "test-svc", CheckType: CheckTypeTTL, TTL: 20, ExpiredTTL: 3, CheckTimeout: 3, Scheme: "http"}
	client, err := NewService("127.0.0.1:8000", conf)
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotEmpty(t, client.GetServiceID())
	assert.NotNil(t, client.GetServiceClient())
	assert.NotNil(t, client.GetRegistration())
}

func TestNewService_HTTP_Success(t *testing.T) {
	srv := fakeConsulServer(t)
	defer srv.Close()

	host := srv.Listener.Addr().String()
	conf := Conf{
		Host: host, Key: "http-svc", CheckType: CheckTypeHttp,
		TTL: 20, ExpiredTTL: 3, CheckTimeout: 3, Scheme: "http",
		CheckHttp: CheckHttpConf{Host: "127.0.0.1", Port: 8001, Scheme: "http", Method: "GET", Path: "/healthz"},
	}
	client, err := NewService("127.0.0.1:8001", conf)
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestNewService_GRPC_Success(t *testing.T) {
	srv := fakeConsulServer(t)
	defer srv.Close()

	host := srv.Listener.Addr().String()
	conf := Conf{Host: host, Key: "grpc-svc", CheckType: CheckTypeGrpc, TTL: 20, ExpiredTTL: 3, CheckTimeout: 3, Scheme: "http"}
	client, err := NewService("127.0.0.1:8002", conf)
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestGetServiceID(t *testing.T) {
	srv := fakeConsulServer(t)
	defer srv.Close()

	host := srv.Listener.Addr().String()
	conf := Conf{Host: host, Key: "my-svc", CheckType: CheckTypeTTL, TTL: 20, ExpiredTTL: 3, CheckTimeout: 3, Scheme: "http"}
	client, err := NewService("127.0.0.1:9000", conf)
	require.NoError(t, err)
	assert.Equal(t, "my-svc-127.0.0.1-9000", client.GetServiceID())
}

func TestDeregisterService(t *testing.T) {
	srv := fakeConsulServer(t)
	defer srv.Close()

	host := srv.Listener.Addr().String()
	conf := Conf{Host: host, Key: "svc", CheckType: CheckTypeTTL, TTL: 20, ExpiredTTL: 3, CheckTimeout: 3, Scheme: "http"}
	client, err := NewService("127.0.0.1:7777", conf)
	require.NoError(t, err)

	err = client.DeregisterService()
	assert.NoError(t, err)
}

func TestWithMonitorFuncs(t *testing.T) {
	srv := fakeConsulServer(t)
	defer srv.Close()

	host := srv.Listener.Addr().String()
	conf := Conf{Host: host, Key: "svc", CheckType: CheckTypeTTL, TTL: 20, ExpiredTTL: 3, CheckTimeout: 3, Scheme: "http"}

	called := make(chan struct{})
	customMonitor := func(cc *CommonClient, stopCh <-chan struct{}) {
		close(called)
		<-stopCh
	}

	client, err := NewService("127.0.0.1:6001", conf, WithMonitorFuncs(customMonitor))
	require.NoError(t, err)

	err = client.RegisterService()
	require.NoError(t, err)

	select {
	case <-called:
		// custom monitor was invoked
	case <-time.After(2 * time.Second):
		t.Fatal("custom monitor was not called")
	}

	client.DeregisterService()
}

func TestRegisterService_TTL(t *testing.T) {
	srv := fakeConsulServer(t)
	defer srv.Close()

	host := srv.Listener.Addr().String()
	conf := Conf{Host: host, Key: "svc", CheckType: CheckTypeTTL, TTL: 20, ExpiredTTL: 3, CheckTimeout: 3, Scheme: "http"}
	client, err := NewService("127.0.0.1:6100", conf)
	require.NoError(t, err)

	err = client.RegisterService()
	assert.NoError(t, err)

	// Cleanup monitors
	client.DeregisterService()
}

// ─────────────────────────────────────────────
// register.go – TTLMonitorLogic / HttpMonitorLogic
// ─────────────────────────────────────────────

func TestTTLMonitorLogic_UpdateTTLSuccess(t *testing.T) {
	// UpdateTTL success path → resets state
	srv := fakeConsulServer(t)
	defer srv.Close()

	host := srv.Listener.Addr().String()
	conf := Conf{Host: host, Key: "svc", CheckType: CheckTypeTTL, TTL: 20, ExpiredTTL: 3, CheckTimeout: 3, Scheme: "http"}
	client, err := NewService("127.0.0.1:6200", conf)
	require.NoError(t, err)
	cc := client.(*CommonClient)

	state := &MonitorState{
		RetryCount:     2,
		BackoffTime:    4 * time.Second,
		MaxRetries:     5,
		MaxBackoffTime: 30 * time.Second,
		OriginalTTL:    5 * time.Second,
		Ticker:         time.NewTicker(time.Hour),
	}
	defer state.Close()

	err = TTLMonitorLogic(cc, state)
	assert.NoError(t, err)
	// On success, RetryCount resets
	assert.Equal(t, 0, state.RetryCount)
}

func TestTTLMonitorLogic_UpdateTTLFail_Registered(t *testing.T) {
	// UpdateTTL fails, but registerServiceHealthStatus says passing → no re-register
	// We need a server that fails UpdateTTL but succeeds for health check.
	var callCount int32
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/agent/check/update/", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.WriteHeader(http.StatusInternalServerError)
	})
	mux.HandleFunc("/v1/agent/service/register", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v1/agent/service/deregister/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v1/agent/check/register", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v1/agent/service/", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{"ID": "svc-127.0.0.1-6201", "Service": "svc"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/v1/health/service/", func(w http.ResponseWriter, r *http.Request) {
		entries := []*api.ServiceEntry{
			{
				Service: &api.AgentService{ID: "svc-127.0.0.1-6201", Service: "svc"},
				Checks:  api.HealthChecks{{Status: api.HealthPassing}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(entries)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	conf := Conf{Host: srv.Listener.Addr().String(), Key: "svc", CheckType: CheckTypeTTL, TTL: 20, ExpiredTTL: 3, CheckTimeout: 3, Scheme: "http"}
	client, err := NewService("127.0.0.1:6201", conf)
	require.NoError(t, err)
	cc := client.(*CommonClient)

	state := &MonitorState{
		RetryCount:     0,
		BackoffTime:    1 * time.Second,
		MaxRetries:     5,
		MaxBackoffTime: 30 * time.Second,
		OriginalTTL:    5 * time.Second,
		Ticker:         time.NewTicker(time.Hour),
	}
	defer state.Close()

	err = TTLMonitorLogic(cc, state)
	assert.NoError(t, err) // registered=true, no re-register needed
}

func TestTTLMonitorLogic_MaxRetriesReached(t *testing.T) {
	// UpdateTTL fails, registerServiceHealthStatus fails, RetryCount >= MaxRetries → reset counter
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/agent/check/update/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	mux.HandleFunc("/v1/agent/service/register", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v1/agent/service/deregister/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v1/agent/check/register", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	// Service returns 500 to simulate health check failure
	mux.HandleFunc("/v1/agent/service/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	mux.HandleFunc("/v1/health/service/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	conf := Conf{Host: srv.Listener.Addr().String(), Key: "svc", CheckType: CheckTypeTTL, TTL: 20, ExpiredTTL: 3, CheckTimeout: 3, Scheme: "http"}
	client, err := NewService("127.0.0.1:6202", conf)
	require.NoError(t, err)
	cc := client.(*CommonClient)

	state := &MonitorState{
		RetryCount:     5, // already at max
		BackoffTime:    30 * time.Second,
		MaxRetries:     5,
		MaxBackoffTime: 30 * time.Second,
		OriginalTTL:    5 * time.Second,
		Ticker:         time.NewTicker(time.Hour),
	}
	defer state.Close()

	err = TTLMonitorLogic(cc, state)
	assert.NoError(t, err)
	// Reset after max retries
	assert.Equal(t, 0, state.RetryCount)
	assert.Equal(t, 1*time.Second, state.BackoffTime)
}

func TestHttpMonitorLogic_Success(t *testing.T) {
	srv := fakeConsulServer(t)
	defer srv.Close()

	host := srv.Listener.Addr().String()
	conf := Conf{
		Host: host, Key: "svc", CheckType: CheckTypeHttp,
		TTL: 20, ExpiredTTL: 3, CheckTimeout: 3, Scheme: "http",
		CheckHttp: CheckHttpConf{Host: "127.0.0.1", Port: 8001, Scheme: "http", Method: "GET", Path: "/healthz"},
	}
	client, err := NewService("127.0.0.1:6300", conf)
	require.NoError(t, err)
	cc := client.(*CommonClient)

	state := &MonitorState{
		RetryCount:     0,
		BackoffTime:    1 * time.Second,
		MaxRetries:     5,
		MaxBackoffTime: 30 * time.Second,
		OriginalTTL:    5 * time.Second,
		Ticker:         time.NewTicker(time.Hour),
	}
	defer state.Close()

	err = HttpMonitorLogic(cc, state)
	assert.NoError(t, err)
}

func TestHttpMonitorLogic_MaxRetriesReached(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/agent/service/register", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v1/agent/service/deregister/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v1/agent/check/register", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	// Health check fails
	mux.HandleFunc("/v1/agent/service/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	mux.HandleFunc("/v1/health/service/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	conf := Conf{
		Host: srv.Listener.Addr().String(), Key: "svc", CheckType: CheckTypeHttp,
		TTL: 20, ExpiredTTL: 3, CheckTimeout: 3, Scheme: "http",
		CheckHttp: CheckHttpConf{Host: "127.0.0.1", Port: 8001, Scheme: "http", Method: "GET", Path: "/healthz"},
	}
	client, err := NewService("127.0.0.1:6301", conf)
	require.NoError(t, err)
	cc := client.(*CommonClient)

	state := &MonitorState{
		RetryCount:     5,
		BackoffTime:    30 * time.Second,
		MaxRetries:     5,
		MaxBackoffTime: 30 * time.Second,
		OriginalTTL:    5 * time.Second,
		Ticker:         time.NewTicker(time.Hour),
	}
	defer state.Close()

	err = HttpMonitorLogic(cc, state)
	assert.NoError(t, err)
	assert.Equal(t, 0, state.RetryCount)
}

func TestTTLCheckMonitorFunc_StopSignal(t *testing.T) {
	srv := fakeConsulServer(t)
	defer srv.Close()

	host := srv.Listener.Addr().String()
	conf := Conf{Host: host, Key: "svc", CheckType: CheckTypeTTL, TTL: 2, ExpiredTTL: 3, CheckTimeout: 3, Scheme: "http"}
	client, err := NewService("127.0.0.1:6400", conf)
	require.NoError(t, err)
	cc := client.(*CommonClient)

	monitorFn := TTLCheckMonitorFunc()
	stopCh := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		monitorFn(cc, stopCh)
	}()

	time.Sleep(50 * time.Millisecond)
	close(stopCh)

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("TTLCheckMonitorFunc did not stop in time")
	}
}

func TestTTLCheckMonitorFunc_ZeroTTL(t *testing.T) {
	// TTL=0 branch: ttlTicker defaults to 1 second
	srv := fakeConsulServer(t)
	defer srv.Close()

	host := srv.Listener.Addr().String()
	conf := Conf{Host: host, Key: "svc", CheckType: CheckTypeTTL, TTL: 1, ExpiredTTL: 3, CheckTimeout: 3, Scheme: "http"}
	// Override TTL to 0 after Validate sets default
	conf.TTL = 0

	client, err := NewService("127.0.0.1:6401", conf)
	require.NoError(t, err)
	cc := client.(*CommonClient)
	cc.consulConf.TTL = 0 // force zero TTL path

	monitorFn := TTLCheckMonitorFunc()
	stopCh := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		monitorFn(cc, stopCh)
	}()

	time.Sleep(50 * time.Millisecond)
	close(stopCh)

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("TTLCheckMonitorFunc (zero TTL) did not stop in time")
	}
}

func TestHttpCheckMonitorFunc_StopSignal(t *testing.T) {
	srv := fakeConsulServer(t)
	defer srv.Close()

	host := srv.Listener.Addr().String()
	conf := Conf{
		Host: host, Key: "svc", CheckType: CheckTypeHttp,
		TTL: 2, ExpiredTTL: 3, CheckTimeout: 3, Scheme: "http",
		CheckHttp: CheckHttpConf{Host: "127.0.0.1", Port: 8001, Scheme: "http", Method: "GET", Path: "/healthz"},
	}
	client, err := NewService("127.0.0.1:6500", conf)
	require.NoError(t, err)
	cc := client.(*CommonClient)

	monitorFn := HttpCheckMonitorFunc()
	stopCh := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		monitorFn(cc, stopCh)
	}()

	time.Sleep(50 * time.Millisecond)
	close(stopCh)

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("HttpCheckMonitorFunc did not stop in time")
	}
}

// ─────────────────────────────────────────────
// resovler.go – byAddressString sort
// ─────────────────────────────────────────────

func TestByAddressString_Sort(t *testing.T) {
	addrs := byAddressString{
		{Addr: "z:9000"},
		{Addr: "a:1000"},
		{Addr: "m:5000"},
	}
	sort.Sort(addrs)
	assert.Equal(t, "a:1000", addrs[0].Addr)
	assert.Equal(t, "m:5000", addrs[1].Addr)
	assert.Equal(t, "z:9000", addrs[2].Addr)
}

func TestByAddressString_Len(t *testing.T) {
	addrs := byAddressString{{Addr: "a"}, {Addr: "b"}}
	assert.Equal(t, 2, addrs.Len())
}

func TestByAddressString_Less(t *testing.T) {
	addrs := byAddressString{{Addr: "a"}, {Addr: "b"}}
	assert.True(t, addrs.Less(0, 1))
	assert.False(t, addrs.Less(1, 0))
}

func TestByAddressString_Swap(t *testing.T) {
	addrs := byAddressString{{Addr: "a"}, {Addr: "b"}}
	addrs.Swap(0, 1)
	assert.Equal(t, "b", addrs[0].Addr)
	assert.Equal(t, "a", addrs[1].Addr)
}

// ─────────────────────────────────────────────
// resovler.go – resolvr.Close / ResolveNow
// ─────────────────────────────────────────────

func TestResolvr_Close(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	r := &resolvr{cancelFunc: cancel}
	r.Close()
	select {
	case <-ctx.Done():
		// expected
	default:
		t.Fatal("context not cancelled after Close()")
	}
}

func TestResolvr_ResolveNow(t *testing.T) {
	r := &resolvr{cancelFunc: func() {}}
	// Should not panic
	r.ResolveNow(resolver.ResolveNowOptions{})
}

// ─────────────────────────────────────────────
// resovler.go – watchConsulService (mock servicer)
// ─────────────────────────────────────────────

type mockServicer struct {
	mu      sync.Mutex
	calls   int
	entries []*api.ServiceEntry
	meta    *api.QueryMeta
	err     error
	// After errCount calls return err, switch to success
	errCount int
}

func (m *mockServicer) Service(service, tag string, passingOnly bool, q *api.QueryOptions) ([]*api.ServiceEntry, *api.QueryMeta, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls++
	if m.errCount > 0 && m.calls <= m.errCount {
		return nil, nil, m.err
	}
	return m.entries, m.meta, nil
}

func TestWatchConsulService_DeliverEntries(t *testing.T) {
	entries := []*api.ServiceEntry{
		{
			Service: &api.AgentService{Address: "10.0.0.1", Port: 8080, Tags: []string{"v1"}},
			Node:    &api.Node{Address: "10.0.0.1"},
		},
		{
			Service: &api.AgentService{Address: "", Port: 9090, Tags: nil},
			Node:    &api.Node{Address: "10.0.0.2"},
		},
	}
	svc := &mockServicer{
		entries: entries,
		meta:    &api.QueryMeta{LastIndex: 1, RequestTime: time.Millisecond},
	}

	tgt := target{
		Service:    "test",
		MaxBackoff: time.Second,
		Near:       "_agent",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	out := make(chan []*consulAddr, 1)
	go watchConsulService(ctx, svc, tgt, out)

	select {
	case addrs := <-out:
		assert.Len(t, addrs, 2)
		// First entry: service address used
		assert.Equal(t, "10.0.0.1", addrs[0].Addr)
		assert.Equal(t, 8080, addrs[0].Port)
		// Second entry: node address fallback
		assert.Equal(t, "10.0.0.2", addrs[1].Addr)
		assert.Equal(t, 9090, addrs[1].Port)
	case <-ctx.Done():
		t.Fatal("timed out waiting for watchConsulService output")
	}
}

func TestWatchConsulService_Limit(t *testing.T) {
	entries := []*api.ServiceEntry{
		{Service: &api.AgentService{Address: "10.0.0.1", Port: 8001}, Node: &api.Node{Address: "10.0.0.1"}},
		{Service: &api.AgentService{Address: "10.0.0.2", Port: 8002}, Node: &api.Node{Address: "10.0.0.2"}},
		{Service: &api.AgentService{Address: "10.0.0.3", Port: 8003}, Node: &api.Node{Address: "10.0.0.3"}},
	}
	svc := &mockServicer{
		entries: entries,
		meta:    &api.QueryMeta{LastIndex: 1},
	}
	tgt := target{Service: "test", MaxBackoff: time.Second, Near: "_agent", Limit: 2}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	out := make(chan []*consulAddr, 1)
	go watchConsulService(ctx, svc, tgt, out)

	select {
	case addrs := <-out:
		assert.Len(t, addrs, 2)
	case <-ctx.Done():
		t.Fatal("timed out")
	}
}

func TestWatchConsulService_ErrorThenSuccess(t *testing.T) {
	entries := []*api.ServiceEntry{
		{Service: &api.AgentService{Address: "10.0.0.1", Port: 8080}, Node: &api.Node{Address: "10.0.0.1"}},
	}
	svc := &mockServicer{
		entries:  entries,
		meta:     &api.QueryMeta{LastIndex: 1},
		err:      fmt.Errorf("consul unavailable"),
		errCount: 1, // first call fails
	}
	tgt := target{Service: "test", MaxBackoff: 50 * time.Millisecond, Near: "_agent"}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	out := make(chan []*consulAddr, 1)
	go watchConsulService(ctx, svc, tgt, out)

	select {
	case addrs := <-out:
		assert.Len(t, addrs, 1)
	case <-ctx.Done():
		t.Fatal("timed out after error-then-success")
	}
}

func TestWatchConsulService_ContextCancel(t *testing.T) {
	// Service blocks forever (channel never receives)
	blockCh := make(chan struct{})
	svc := &mockServicer{
		entries: []*api.ServiceEntry{},
		meta:    &api.QueryMeta{LastIndex: 1},
	}
	_ = blockCh

	tgt := target{Service: "test", MaxBackoff: time.Second, Near: "_agent"}
	ctx, cancel := context.WithCancel(context.Background())
	out := make(chan []*consulAddr)

	go watchConsulService(ctx, svc, tgt, out)

	// Let one result through
	go func() {
		select {
		case <-out:
		case <-ctx.Done():
		}
	}()

	cancel()
	// No assertion needed — just verify no goroutine leak / panic
	time.Sleep(100 * time.Millisecond)
}

// ─────────────────────────────────────────────
// resovler.go – populateEndpoints (mock clientConn)
// ─────────────────────────────────────────────

type mockClientConn struct {
	resolver.ClientConn // embed to satisfy future interface additions
	mu     sync.Mutex
	states []resolver.State
}

func (m *mockClientConn) UpdateState(s resolver.State) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.states = append(m.states, s)
	return nil
}

func (m *mockClientConn) ReportError(error)             {}
func (m *mockClientConn) NewAddress([]resolver.Address) {}

func TestPopulateEndpoints_Basic(t *testing.T) {
	cc := &mockClientConn{}
	input := make(chan []*consulAddr, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go populateEndpoints(ctx, cc, input)

	input <- []*consulAddr{
		{Addr: "10.0.0.2", Port: 9000, Tags: []string{"v2"}},
		{Addr: "10.0.0.1", Port: 8000, Tags: nil},
		{Addr: "10.0.0.2", Port: 9000, Tags: []string{"v2"}}, // duplicate
	}

	time.Sleep(100 * time.Millisecond)
	cancel()
	time.Sleep(50 * time.Millisecond)

	cc.mu.Lock()
	defer cc.mu.Unlock()
	require.Len(t, cc.states, 1)
	// Duplicates collapsed → 2 unique addresses
	assert.Len(t, cc.states[0].Addresses, 2)
	// Sorted: 10.0.0.1 < 10.0.0.2
	assert.Equal(t, "10.0.0.1:8000", cc.states[0].Addresses[0].Addr)
	assert.Equal(t, "10.0.0.2:9000", cc.states[0].Addresses[1].Addr)
}

func TestPopulateEndpoints_TagsSet(t *testing.T) {
	cc := &mockClientConn{}
	input := make(chan []*consulAddr, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go populateEndpoints(ctx, cc, input)

	input <- []*consulAddr{
		{Addr: "10.0.0.1", Port: 8080, Tags: []string{"tag1", "tag2"}},
	}

	time.Sleep(100 * time.Millisecond)
	cancel()
	time.Sleep(50 * time.Millisecond)

	cc.mu.Lock()
	defer cc.mu.Unlock()
	require.Len(t, cc.states, 1)
	require.Len(t, cc.states[0].Addresses, 1)
	// Tags attribute should be set (not nil)
	assert.NotNil(t, cc.states[0].Addresses[0].Attributes)
}

func TestPopulateEndpoints_ContextCancel(t *testing.T) {
	cc := &mockClientConn{}
	input := make(chan []*consulAddr)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	done := make(chan struct{})
	go func() {
		populateEndpoints(ctx, cc, input)
		close(done)
	}()

	select {
	case <-done:
		// populateEndpoints returned on ctx cancel
	case <-time.After(2 * time.Second):
		t.Fatal("populateEndpoints did not exit on context cancel")
	}
}

// ─────────────────────────────────────────────
// builder.go – Build() with bad URL (wrong scheme)
// ─────────────────────────────────────────────

func TestBuilder_Scheme(t *testing.T) {
	b := &builder{}
	assert.Equal(t, schemeName, b.Scheme())
}

func TestBuilder_Build_BadURL(t *testing.T) {
	b := &builder{}
	// wrong scheme → parseURL returns error
	u, _ := url.Parse("http://localhost:8500/svc")
	target := resolver.Target{URL: *u}
	_, err := b.Build(target, &mockClientConn{}, resolver.BuildOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Wrong consul URL")
}

// ─────────────────────────────────────────────
// Additional coverage tests
// ─────────────────────────────────────────────

func TestBuilder_Build_ValidURL(t *testing.T) {
	// api.NewClient never fails; watchConsulService and populateEndpoints
	// are started as goroutines — the resolver is returned.
	b := &builder{}
	u, _ := url.Parse("consul://localhost:8500/my-service")
	tgt := resolver.Target{URL: *u}
	r, err := b.Build(tgt, &mockClientConn{}, resolver.BuildOptions{})
	require.NoError(t, err)
	require.NotNil(t, r)
	// Close via the resolvr to cancel goroutines
	r.Close()
}

func TestMustNewService_Success(t *testing.T) {
	srv := fakeConsulServer(t)
	defer srv.Close()

	host := srv.Listener.Addr().String()
	conf := Conf{Host: host, Key: "svc", CheckType: CheckTypeTTL, TTL: 20, ExpiredTTL: 3, CheckTimeout: 3, Scheme: "http"}
	client := MustNewService("127.0.0.1:9500", conf)
	require.NotNil(t, client)
}

func TestMustNewService_Panics(t *testing.T) {
	// Set ExitOnFatal=false so logx.Must panics instead of os.Exit
	logx.ExitOnFatal.Set(false)
	t.Cleanup(func() { logx.ExitOnFatal.Set(true) })

	defer func() {
		r := recover()
		require.NotNil(t, r, "expected MustNewService to panic on invalid conf")
	}()

	// Empty host → Validate fails → logx.Must panics
	MustNewService("127.0.0.1:9501", Conf{Key: "svc"})
}

func TestResolvr_ResolveNow_NoOp(t *testing.T) {
	r := &resolvr{cancelFunc: func() {}}
	// Simply verify no panic and it returns immediately
	r.ResolveNow(resolver.ResolveNowOptions{})
}

func TestTTLMonitorLogic_ReregistrationRetry(t *testing.T) {
	// UpdateTTL fails, health check fails, RetryCount < MaxRetries → re-register attempt
	var registerCalled int32
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/agent/check/update/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	mux.HandleFunc("/v1/agent/service/register", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&registerCalled, 1)
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v1/agent/service/deregister/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v1/agent/check/register", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	// Health check fails → not registered
	mux.HandleFunc("/v1/agent/service/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	mux.HandleFunc("/v1/health/service/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	conf := Conf{Host: srv.Listener.Addr().String(), Key: "svc", CheckType: CheckTypeTTL, TTL: 20, ExpiredTTL: 3, CheckTimeout: 3, Scheme: "http"}
	client, err := NewService("127.0.0.1:6203", conf)
	require.NoError(t, err)
	cc := client.(*CommonClient)

	state := &MonitorState{
		RetryCount:     0,
		BackoffTime:    1 * time.Second,
		MaxRetries:     5,
		MaxBackoffTime: 30 * time.Second,
		OriginalTTL:    5 * time.Second,
		Ticker:         time.NewTicker(time.Hour),
	}
	defer state.Close()

	// re-register succeeds → RetryCount resets
	err = TTLMonitorLogic(cc, state)
	assert.NoError(t, err)
	assert.Equal(t, 0, state.RetryCount) // reset after successful re-register
}

func TestHttpMonitorLogic_ReregistrationRetry(t *testing.T) {
	// Health check fails, RetryCount < MaxRetries → re-register attempt succeeds
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/agent/service/register", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v1/agent/service/deregister/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v1/agent/check/register", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v1/agent/service/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	mux.HandleFunc("/v1/health/service/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	conf := Conf{
		Host: srv.Listener.Addr().String(), Key: "svc", CheckType: CheckTypeHttp,
		TTL: 20, ExpiredTTL: 3, CheckTimeout: 3, Scheme: "http",
		CheckHttp: CheckHttpConf{Host: "127.0.0.1", Port: 8001, Scheme: "http", Method: "GET", Path: "/healthz"},
	}
	client, err := NewService("127.0.0.1:6302", conf)
	require.NoError(t, err)
	cc := client.(*CommonClient)

	state := &MonitorState{
		RetryCount:     0,
		BackoffTime:    1 * time.Second,
		MaxRetries:     5,
		MaxBackoffTime: 30 * time.Second,
		OriginalTTL:    5 * time.Second,
		Ticker:         time.NewTicker(time.Hour),
	}
	defer state.Close()

	err = HttpMonitorLogic(cc, state)
	assert.NoError(t, err)
	assert.Equal(t, 0, state.RetryCount)
}

func TestSetRegisterServiceHealthStatus_UnknownType(t *testing.T) {
	srv := fakeConsulServer(t)
	defer srv.Close()

	host := srv.Listener.Addr().String()
	conf := Conf{Host: host, Key: "svc", CheckType: CheckTypeTTL, TTL: 20, ExpiredTTL: 3, CheckTimeout: 3, Scheme: "http"}
	client, err := NewService("127.0.0.1:6600", conf)
	require.NoError(t, err)
	cc := client.(*CommonClient)

	// Force unknown check type after construction
	cc.consulConf.CheckType = "unknown"
	err = cc.setRegisterServiceHealthStatus(api.HealthPassing)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown check type")
}

func TestRegisterServiceMonitors_UnknownType(t *testing.T) {
	srv := fakeConsulServer(t)
	defer srv.Close()

	host := srv.Listener.Addr().String()
	conf := Conf{Host: host, Key: "svc", CheckType: CheckTypeTTL, TTL: 20, ExpiredTTL: 3, CheckTimeout: 3, Scheme: "http"}
	client, err := NewService("127.0.0.1:6601", conf)
	require.NoError(t, err)
	cc := client.(*CommonClient)

	// Force unknown check type, clear existing monitors
	cc.consulConf.CheckType = "unknown"
	cc.monitorFuncs = nil
	err = cc.registerServiceMonitors()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown check type")
}

func TestTTLCheckMonitorFunc_TickerFires(t *testing.T) {
	// Start TTLCheckMonitorFunc with TTL=2, let the ticker fire once, then stop.
	srv := fakeConsulServer(t)
	defer srv.Close()

	host := srv.Listener.Addr().String()
	conf := Conf{Host: host, Key: "svc", CheckType: CheckTypeTTL, TTL: 2, ExpiredTTL: 3, CheckTimeout: 3, Scheme: "http"}
	client, err := NewService("127.0.0.1:6700", conf)
	require.NoError(t, err)
	cc := client.(*CommonClient)

	monitorFn := TTLCheckMonitorFunc()
	stopCh := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		monitorFn(cc, stopCh)
	}()

	// Wait long enough for the ticker to fire at least once (TTL-1 = 1s)
	time.Sleep(1200 * time.Millisecond)
	close(stopCh)

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("monitor did not stop")
	}
}

// ─────────────────────────────────────────────
// Additional gap-filling tests
// ─────────────────────────────────────────────

// TTLCheckMonitorFunc: TTL=1 triggers ttlTicker < time.Second branch (reset to 1s)
func TestTTLCheckMonitorFunc_TTLOne(t *testing.T) {
	srv := fakeConsulServer(t)
	defer srv.Close()

	host := srv.Listener.Addr().String()
	conf := Conf{Host: host, Key: "svc", CheckType: CheckTypeTTL, TTL: 1, ExpiredTTL: 3, CheckTimeout: 3, Scheme: "http"}
	client, err := NewService("127.0.0.1:6800", conf)
	require.NoError(t, err)
	cc := client.(*CommonClient)

	monitorFn := TTLCheckMonitorFunc()
	stopCh := make(chan struct{})
	done := make(chan struct{})
	go func() { defer close(done); monitorFn(cc, stopCh) }()
	time.Sleep(50 * time.Millisecond)
	close(stopCh)
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("monitor did not stop")
	}
}

// HttpCheckMonitorFunc: zero TTL branch (ttlTicker defaults to 1s)
func TestHttpCheckMonitorFunc_ZeroTTL(t *testing.T) {
	srv := fakeConsulServer(t)
	defer srv.Close()

	host := srv.Listener.Addr().String()
	conf := Conf{
		Host: host, Key: "svc", CheckType: CheckTypeHttp,
		TTL: 1, ExpiredTTL: 3, CheckTimeout: 3, Scheme: "http",
		CheckHttp: CheckHttpConf{Host: "127.0.0.1", Port: 8001, Scheme: "http", Method: "GET", Path: "/healthz"},
	}
	client, err := NewService("127.0.0.1:6801", conf)
	require.NoError(t, err)
	cc := client.(*CommonClient)
	cc.consulConf.TTL = 0 // force zero-TTL path

	monitorFn := HttpCheckMonitorFunc()
	stopCh := make(chan struct{})
	done := make(chan struct{})
	go func() { defer close(done); monitorFn(cc, stopCh) }()
	time.Sleep(50 * time.Millisecond)
	close(stopCh)
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("monitor did not stop")
	}
}

// TTLMonitorLogic: re-register fails → retry count increments and backoff doubles
func TestTTLMonitorLogic_ReregistrationFails(t *testing.T) {
	var fail atomic.Bool
	mux := http.NewServeMux()
	// UpdateTTL endpoint: always 500 so TTL update fails
	mux.HandleFunc("/v1/agent/check/update/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	// Register: OK during setup, then 500 when re-register is attempted
	mux.HandleFunc("/v1/agent/service/register", func(w http.ResponseWriter, r *http.Request) {
		if fail.Load() {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	})
	mux.HandleFunc("/v1/agent/service/deregister/", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("/v1/agent/check/register", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	// Agent service lookup: 500 → getRegisterServiceHealthStatus fails → registered=false
	mux.HandleFunc("/v1/agent/service/", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusInternalServerError) })
	mux.HandleFunc("/v1/health/service/", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusInternalServerError) })
	srv := httptest.NewServer(mux)
	defer srv.Close()

	conf := Conf{Host: srv.Listener.Addr().String(), Key: "svc", CheckType: CheckTypeTTL, TTL: 20, ExpiredTTL: 3, CheckTimeout: 3, Scheme: "http"}
	client, err := NewService("127.0.0.1:6802", conf)
	require.NoError(t, err)
	cc := client.(*CommonClient)

	// Now flip to fail mode so re-register returns 500
	fail.Store(true)

	state := &MonitorState{
		RetryCount: 0, BackoffTime: 1 * time.Second,
		MaxRetries: 5, MaxBackoffTime: 30 * time.Second,
		OriginalTTL: 5 * time.Second,
		Ticker:      time.NewTicker(time.Hour),
	}
	defer state.Close()

	err = TTLMonitorLogic(cc, state)
	require.Error(t, err) // re-register failed
	assert.Equal(t, 1, state.RetryCount)
	assert.Equal(t, 2*time.Second, state.BackoffTime) // doubled
}

// HttpMonitorLogic: re-register fails → retry count increments
func TestHttpMonitorLogic_ReregistrationFails(t *testing.T) {
	var fail atomic.Bool
	mux := http.NewServeMux()
	// Register: OK during setup, 500 on re-register attempt
	mux.HandleFunc("/v1/agent/service/register", func(w http.ResponseWriter, r *http.Request) {
		if fail.Load() {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	})
	mux.HandleFunc("/v1/agent/service/deregister/", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("/v1/agent/check/register", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	// Agent service lookup: 500 → registered=false
	mux.HandleFunc("/v1/agent/service/", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusInternalServerError) })
	mux.HandleFunc("/v1/health/service/", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusInternalServerError) })
	srv := httptest.NewServer(mux)
	defer srv.Close()

	conf := Conf{
		Host: srv.Listener.Addr().String(), Key: "svc", CheckType: CheckTypeHttp,
		TTL: 20, ExpiredTTL: 3, CheckTimeout: 3, Scheme: "http",
		CheckHttp: CheckHttpConf{Host: "127.0.0.1", Port: 8001, Scheme: "http", Method: "GET", Path: "/healthz"},
	}
	client, err := NewService("127.0.0.1:6803", conf)
	require.NoError(t, err)
	cc := client.(*CommonClient)

	// Now flip to fail mode
	fail.Store(true)

	state := &MonitorState{
		RetryCount: 0, BackoffTime: 1 * time.Second,
		MaxRetries: 5, MaxBackoffTime: 30 * time.Second,
		OriginalTTL: 5 * time.Second,
		Ticker:      time.NewTicker(time.Hour),
	}
	defer state.Close()

	err = HttpMonitorLogic(cc, state)
	require.Error(t, err)
	assert.Equal(t, 1, state.RetryCount)
	assert.Equal(t, 2*time.Second, state.BackoffTime)
}

// registerServiceHealthStatus: status mismatch returns error
func TestRegisterServiceHealthStatus_Mismatch(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/agent/service/register", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("/v1/agent/service/deregister/", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("/v1/agent/check/register", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("/v1/agent/service/", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{"ID": "svc-127.0.0.1-6804", "Service": "svc"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/v1/health/service/", func(w http.ResponseWriter, r *http.Request) {
		entries := []*api.ServiceEntry{
			{
				Service: &api.AgentService{ID: "svc-127.0.0.1-6804", Service: "svc"},
				Checks:  api.HealthChecks{{Status: api.HealthCritical}}, // critical, not passing
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(entries)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	conf := Conf{Host: srv.Listener.Addr().String(), Key: "svc", CheckType: CheckTypeTTL, TTL: 20, ExpiredTTL: 3, CheckTimeout: 3, Scheme: "http"}
	client, err := NewService("127.0.0.1:6804", conf)
	require.NoError(t, err)
	cc := client.(*CommonClient)

	ok, err := cc.registerServiceHealthStatus(api.HealthPassing)
	require.Error(t, err)
	assert.False(t, ok)
	assert.Contains(t, err.Error(), "not passing")
}

// getRegisterServiceHealthStatus: service ID not found in health entries
func TestGetRegisterServiceHealthStatus_NotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/agent/service/register", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("/v1/agent/service/deregister/", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("/v1/agent/check/register", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("/v1/agent/service/", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{"ID": "svc-127.0.0.1-6805", "Service": "svc"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/v1/health/service/", func(w http.ResponseWriter, r *http.Request) {
		// Return entries with a DIFFERENT service ID
		entries := []*api.ServiceEntry{
			{
				Service: &api.AgentService{ID: "other-service-id", Service: "svc"},
				Checks:  api.HealthChecks{{Status: api.HealthPassing}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(entries)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	conf := Conf{Host: srv.Listener.Addr().String(), Key: "svc", CheckType: CheckTypeTTL, TTL: 20, ExpiredTTL: 3, CheckTimeout: 3, Scheme: "http"}
	client, err := NewService("127.0.0.1:6805", conf)
	require.NoError(t, err)
	cc := client.(*CommonClient)

	_, err = cc.getRegisterServiceHealthStatus()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// registerServiceMonitors: already has monitor funcs → skips default assignment (RLock short-circuit)
func TestRegisterServiceMonitors_AlreadyHasFuncs(t *testing.T) {
	srv := fakeConsulServer(t)
	defer srv.Close()

	host := srv.Listener.Addr().String()
	conf := Conf{Host: host, Key: "svc", CheckType: CheckTypeTTL, TTL: 20, ExpiredTTL: 3, CheckTimeout: 3, Scheme: "http"}

	called := make(chan struct{}, 1)
	customMonitor := func(cc *CommonClient, stopCh <-chan struct{}) {
		called <- struct{}{}
		<-stopCh
	}

	client, err := NewService("127.0.0.1:6806", conf, WithMonitorFuncs(customMonitor))
	require.NoError(t, err)
	cc := client.(*CommonClient)

	// monitorFuncs already set via WithMonitorFuncs; registerServiceMonitors should skip default assignment
	err = cc.registerServiceMonitors()
	require.NoError(t, err)

	// Stop all monitors to avoid goroutine leak
	cc.stopAllMonitors()
}

// parseURL: mapping.UnmarshalKey fails on invalid param type
func TestParseURL_UnmarshalError(t *testing.T) {
	// 'wait' expects a duration; passing a non-parseable value causes UnmarshalKey to fail
	u := mustParseURL("consul://localhost:8500/svc?wait=not-a-duration")
	_, err := parseURL(u)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Malformed URL parameters")
}

// setRegisterServiceHealthStatus gRPC CheckRegister failure path
func TestSetRegisterServiceHealthStatus_GrpcCheckRegisterFail(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/agent/service/register", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("/v1/agent/service/deregister/", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("/v1/agent/check/register", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	conf := Conf{Host: srv.Listener.Addr().String(), Key: "svc", CheckType: CheckTypeGrpc, TTL: 20, ExpiredTTL: 3, CheckTimeout: 3, Scheme: "http"}
	client, err := NewService("127.0.0.1:6807", conf)
	require.NoError(t, err)
	cc := client.(*CommonClient)

	err = cc.setRegisterServiceHealthStatus(api.HealthPassing)
	require.Error(t, err)
}
// ─────────────────────────────────────────────
// register.go – RegisterService error paths
// ─────────────────────────────────────────────

func TestRegisterService_RegisterWithPassingHealthFails(t *testing.T) {
	// ServiceRegister returns 500 → registerServiceWithPassingHealth fails → RegisterService returns error
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/agent/service/register", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	mux.HandleFunc("/v1/agent/service/deregister/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/v1/agent/check/register", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	apiClient, err := api.NewClient(&api.Config{
		Scheme:  "http",
		Address: srv.Listener.Addr().String(),
	})
	require.NoError(t, err)

	conf := Conf{Host: srv.Listener.Addr().String(), Key: "svc", CheckType: CheckTypeTTL, TTL: 20, ExpiredTTL: 3, CheckTimeout: 3, Scheme: "http"}
	cc := &CommonClient{
		serviceId:    "svc-127.0.0.1-7000",
		serviceHost:  "127.0.0.1",
		servicePort:  7000,
		consulConf:   conf,
		apiClient:    apiClient,
		monitorFuncs: make([]MonitorFunc, 0),
	}
	require.NoError(t, cc.clientRegistration())

	err = cc.RegisterService()
	require.Error(t, err)
}

func TestRegisterService_MonitorsFail(t *testing.T) {
	srv := fakeConsulServer(t)
	defer srv.Close()

	host := srv.Listener.Addr().String()
	conf := Conf{Host: host, Key: "svc", CheckType: CheckTypeTTL, TTL: 20, ExpiredTTL: 3, CheckTimeout: 3, Scheme: "http"}
	client, err := NewService("127.0.0.1:7001", conf)
	require.NoError(t, err)
	cc := client.(*CommonClient)

	// Force monitors to fail: unknown check type + no pre-existing funcs
	cc.consulConf.CheckType = "unknown"
	cc.monitorFuncs = nil

	err = cc.RegisterService()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown check type")
}

func TestRegisterService_ShutdownListenerFires(t *testing.T) {
	// proc.Shutdown() triggers the AddShutdownListener closure, covering lines 130-136.
	srv := fakeConsulServer(t)
	defer srv.Close()

	host := srv.Listener.Addr().String()
	conf := Conf{Host: host, Key: "svc", CheckType: CheckTypeTTL, TTL: 20, ExpiredTTL: 3, CheckTimeout: 3, Scheme: "http"}
	client, err := NewService("127.0.0.1:7002", conf)
	require.NoError(t, err)

	// RegisterService registers the shutdown listener; capture waitForCalled via the return value.
	// We wrap RegisterService and intercept the listener by calling proc.Shutdown() after.
	err = client.RegisterService()
	require.NoError(t, err)

	// proc.Shutdown() calls all registered shutdown listeners synchronously (test-only API).
	// This exercises the closure: stopAllMonitors + deleteRegisterService.
	proc.Shutdown()
}
func TestHttpCheckMonitorFunc_TickerFires(t *testing.T) {
	// Start HttpCheckMonitorFunc with TTL=2, let the ticker fire once, then stop.
	srv := fakeConsulServer(t)
	defer srv.Close()

	host := srv.Listener.Addr().String()
	conf := Conf{
		Host: host, Key: "svc", CheckType: CheckTypeHttp,
		TTL: 2, ExpiredTTL: 3, CheckTimeout: 3, Scheme: "http",
		CheckHttp: CheckHttpConf{Host: "127.0.0.1", Port: 8001, Scheme: "http", Method: "GET", Path: "/healthz"},
	}
	client, err := NewService("127.0.0.1:6701", conf)
	require.NoError(t, err)
	cc := client.(*CommonClient)

	monitorFn := HttpCheckMonitorFunc()
	stopCh := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		monitorFn(cc, stopCh)
	}()

	time.Sleep(1200 * time.Millisecond)
	close(stopCh)

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("monitor did not stop")
	}
}
