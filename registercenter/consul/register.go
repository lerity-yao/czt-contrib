package consul

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/netx"
	"github.com/zeromicro/go-zero/core/proc"
)

type (
	Client interface {
		RegisterService() error
		DeregisterService() error
		GetServiceID() string
		GetRegistration() *api.AgentServiceRegistration
		GetServiceClient() *api.Client
	}

	CommonClient struct {
		registration   *api.AgentServiceRegistration
		apiClient      *api.Client
		serviceId      string
		serviceHost    string
		servicePort    int
		consulConf     Conf
		monitorFuncs   []MonitorFunc
		monitorMutex   sync.RWMutex
		stopMonitorChs []chan struct{}
		stopChMutex    sync.Mutex
	}

	MonitorState struct {
		RetryCount     int
		BackoffTime    time.Duration
		MaxBackoffTime time.Duration
		MaxRetries     int
		Ticker         *time.Ticker
		OriginalTTL    time.Duration // 保存原始的TTL值，用于重置
		Mutex          sync.RWMutex
	}

	MonitorFunc func(cc *CommonClient, stopChan <-chan struct{})

	ServiceOption func(*CommonClient)
)

func MustNewService(listenOn string, c Conf, opts ...ServiceOption) Client {
	service, err := NewService(listenOn, c, opts...)
	if err != nil {
		logx.Must(err)
	}
	return service
}

func NewService(listenOn string, c Conf, opts ...ServiceOption) (Client, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	pubListenOn := figureOutListenOn(listenOn)

	host, ports, err := net.SplitHostPort(pubListenOn)
	if err != nil {
		return nil, fmt.Errorf("failed parsing address error: %v", err)
	}
	port, _ := strconv.ParseUint(ports, 10, 16)

	client, err := api.NewClient(&api.Config{Scheme: c.Scheme, Address: c.Host, Token: c.Token})
	if err != nil {
		return nil, fmt.Errorf("create consul client error: %v", err)
	}

	service := &CommonClient{
		registration: nil,
		apiClient:    client,
		serviceId:    fmt.Sprintf("%s-%s-%d", c.Key, host, port),
		serviceHost:  host,
		servicePort:  int(port),
		consulConf:   c,
		monitorFuncs: make([]MonitorFunc, 0),
	}

	for _, option := range opts {
		option(service)
	}

	err = service.clientRegistration()
	if err != nil {
		return nil, err
	}

	return service, nil
}

func (cc *CommonClient) RegisterService() error {
	err := cc.registerServiceWithPassingHealth()
	if err != nil {
		return err
	}

	err = cc.registerServiceMonitors()
	if err != nil {
		return err
	}

	proc.AddShutdownListener(func() {
		cc.stopAllMonitors()
		err := cc.deleteRegisterService()
		if err != nil {
			logx.Errorf("deregister service %s error: %s", cc.serviceId, err.Error())
		} else {
			logx.Infof("Service %s deregistered successfully", cc.serviceId)
		}
	})
	return nil
}

func (cc *CommonClient) DeregisterService() error {
	cc.stopAllMonitors()
	return cc.deleteRegisterService()
}

func (cc *CommonClient) GetServiceID() string {
	return cc.serviceId
}

func (cc *CommonClient) GetServiceClient() *api.Client {
	return cc.apiClient
}
func (cc *CommonClient) GetRegistration() *api.AgentServiceRegistration {
	return cc.registration
}

func (cc *CommonClient) registerServiceWithPassingHealth() error {

	_ = cc.deleteRegisterService()
	time.Sleep(100 * time.Millisecond)

	err := cc.registerService()
	if err != nil {
		return err
	}

	err = cc.setRegisterServiceHealthStatus(api.HealthPassing)
	if err != nil {
		_ = cc.deleteRegisterService()
		return fmt.Errorf("register service %s id %s check error: %v",
			cc.registration.Name, cc.registration.ID, err)
	}

	logx.Infof("Service %s id %s registered successfully", cc.registration.Name, cc.registration.ID)

	return nil
}

func (cc *CommonClient) registerService() error {
	err := cc.apiClient.Agent().ServiceRegister(cc.registration)
	if err != nil {
		return fmt.Errorf("register service %s id %s host to consul error: %s",
			cc.registration.Name, cc.registration.ID, err.Error())
	}
	return nil
}

func (cc *CommonClient) clientRegistration() error {
	reg := &api.AgentServiceRegistration{
		ID:      cc.serviceId,       // Service node name
		Name:    cc.consulConf.Key,  // Service name
		Tags:    cc.consulConf.Tag,  // Tags, can be empty
		Meta:    cc.consulConf.Meta, // Meta, can be empty
		Port:    cc.servicePort,     // Service port
		Address: cc.serviceHost,     // Service IP
	}

	switch cc.consulConf.CheckType {
	case CheckTypeTTL:
		reg.Checks = []*api.AgentServiceCheck{
			{
				CheckID:                        cc.serviceId,                          // Service node name
				TTL:                            fmt.Sprintf("%ds", cc.consulConf.TTL), // Health check interval
				Status:                         api.HealthPassing,
				DeregisterCriticalServiceAfter: fmt.Sprintf("%ds", cc.consulConf.TTL*cc.consulConf.ExpiredTTL), // Deregistration time
			},
		}

	case CheckTypeGrpc:
	case CheckTypeHttp:
		// todo 可以考虑混合健康检查，例如TTL和HTTP
		httpCheckHost := figureOutListenOn(fmt.Sprintf("%s:%d", cc.consulConf.CheckHttp.Host, cc.consulConf.CheckHttp.Port))
		reg.Checks = []*api.AgentServiceCheck{
			{
				CheckID:                        cc.serviceId,                                                                                          // Service node name
				HTTP:                           fmt.Sprintf("%s://%s%s", cc.consulConf.CheckHttp.Scheme, httpCheckHost, cc.consulConf.CheckHttp.Path), // health check url
				Method:                         cc.consulConf.CheckHttp.Method,                                                                        // health check method
				Interval:                       fmt.Sprintf("%ds", cc.consulConf.TTL),                                                                 // health check interval
				Timeout:                        fmt.Sprintf("%ds", cc.consulConf.CheckTimeout),                                                        // health check timeout
				DeregisterCriticalServiceAfter: fmt.Sprintf("%ds", cc.consulConf.TTL*cc.consulConf.ExpiredTTL),
				Status:                         api.HealthPassing,
			},
		}
	default:
		return fmt.Errorf("unknown check type: %s", cc.consulConf.CheckType)
	}
	cc.registration = reg
	return nil
}

func (cc *CommonClient) deleteRegisterService() error {
	err := cc.apiClient.Agent().ServiceDeregister(cc.serviceId)
	return err
}

func (cc *CommonClient) setRegisterServiceHealthStatus(status string) error {
	switch cc.consulConf.CheckType {
	case CheckTypeTTL:
		check := api.AgentServiceCheck{
			TTL:                            fmt.Sprintf("%ds", cc.consulConf.TTL),
			Status:                         status,
			DeregisterCriticalServiceAfter: fmt.Sprintf("%ds", cc.consulConf.TTL*cc.consulConf.ExpiredTTL),
		}
		return cc.apiClient.Agent().CheckRegister(&api.AgentCheckRegistration{
			ID:                cc.registration.ID,
			Name:              cc.registration.Name,
			ServiceID:         cc.registration.ID,
			AgentServiceCheck: check,
		})
	case CheckTypeHttp:
		httpCheckHost := figureOutListenOn(cc.consulConf.CheckHttp.Host)
		check := api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("%s://%s%s", cc.consulConf.CheckHttp.Scheme, httpCheckHost, cc.consulConf.CheckHttp.Path), // health check url
			Method:                         cc.consulConf.CheckHttp.Method,                                                                        // health check method
			Interval:                       fmt.Sprintf("%ds", cc.consulConf.TTL),                                                                 // health check interval
			Timeout:                        fmt.Sprintf("%ds", cc.consulConf.CheckTimeout),                                                        // health check timeout
			DeregisterCriticalServiceAfter: fmt.Sprintf("%ds", cc.consulConf.TTL*cc.consulConf.ExpiredTTL),                                        // Deregistration time
			Status:                         status,
		}
		return cc.apiClient.Agent().CheckRegister(&api.AgentCheckRegistration{
			ID:                cc.registration.ID,
			Name:              cc.registration.Name,
			ServiceID:         cc.registration.ID,
			AgentServiceCheck: check,
		})
	case CheckTypeGrpc:
	default:
		return fmt.Errorf("unknown check type: %s", cc.consulConf.CheckType)
	}

	return nil
}

func (cc *CommonClient) getRegisterServiceHealthStatus() (string, error) {
	service, _, err := cc.apiClient.Agent().Service(cc.serviceId, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get service %s: %v", cc.serviceId, err)
	}

	serviceEntries, _, err := cc.apiClient.Health().Service(service.Service, "", false, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get service %s health: %v", cc.serviceId, err)
	}

	for _, entry := range serviceEntries {
		if entry.Service.ID == cc.serviceId {
			return entry.Checks.AggregatedStatus(), nil
		}
	}
	return "", fmt.Errorf("service %s not found in health check results", cc.serviceId)
}

func (cc *CommonClient) registerServiceHealthStatus(status string) (bool, error) {
	ss, err := cc.getRegisterServiceHealthStatus()
	if err != nil {
		return false, err
	}

	if ss != status {
		return false, fmt.Errorf("service status is %s, not %s", ss, status)
	}

	return true, nil
}

func (cc *CommonClient) registerServiceMonitors() error {
	cc.monitorMutex.RLock()
	hasNoFuncs := len(cc.monitorFuncs) == 0
	cc.monitorMutex.RUnlock()

	if hasNoFuncs {
		cc.monitorMutex.Lock()
		if len(cc.monitorFuncs) == 0 {
			switch cc.consulConf.CheckType {
			case CheckTypeTTL:
				cc.monitorFuncs = append(cc.monitorFuncs, TTLCheckMonitorFunc())
			case CheckTypeGrpc:
				cc.monitorFuncs = append(cc.monitorFuncs, HttpCheckMonitorFunc())
			case CheckTypeHttp:
				cc.monitorFuncs = append(cc.monitorFuncs, HttpCheckMonitorFunc())
			default:
				return fmt.Errorf("unknown check type: %s", cc.consulConf.CheckType)
			}

		}
		cc.monitorMutex.Unlock()
	}

	cc.monitorMutex.RLock()
	funcsCopy := make([]MonitorFunc, len(cc.monitorFuncs))
	copy(funcsCopy, cc.monitorFuncs)
	cc.monitorMutex.RUnlock()
	cc.startMonitors(funcsCopy)
	return nil
}

func (cc *CommonClient) startMonitors(funcs []MonitorFunc) {
	cc.stopChMutex.Lock()
	defer cc.stopChMutex.Unlock()

	for _, monitorFunc := range funcs {
		stopCh := make(chan struct{})
		cc.stopMonitorChs = append(cc.stopMonitorChs, stopCh)
		go monitorFunc(cc, stopCh)
	}
}

func (cc *CommonClient) stopAllMonitors() {
	cc.stopChMutex.Lock()
	defer cc.stopChMutex.Unlock()

	for _, stopCh := range cc.stopMonitorChs {
		close(stopCh)
	}
	cc.stopMonitorChs = make([]chan struct{}, 0)
}

func (s *MonitorState) Close() {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	if s.Ticker != nil {
		s.Ticker.Stop()
		s.Ticker = nil
	}
}

func TTLCheckMonitorFunc() MonitorFunc {
	return func(cc *CommonClient, stopCh <-chan struct{}) {
		var ttlTicker time.Duration
		if cc.consulConf.TTL > 0 {
			ttlTicker = time.Duration(cc.consulConf.TTL-1) * time.Second
			if ttlTicker < time.Second {
				ttlTicker = time.Second
			}
		} else {
			ttlTicker = 1 * time.Second
		}

		state := &MonitorState{
			RetryCount:     0,
			BackoffTime:    1 * time.Second,
			MaxRetries:     5,
			MaxBackoffTime: 30 * time.Second,
			OriginalTTL:    ttlTicker,
			Ticker:         time.NewTicker(ttlTicker),
		}
		defer state.Close()

		for {
			select {
			case <-state.Ticker.C:
				err := TTLMonitorLogic(cc, state)
				if err != nil {
					logx.Errorf("Monitor function error for service %s: %v", cc.serviceId, err)
				}
			case <-stopCh:
				logx.Infof("Service monitor for %s stopped gracefully", cc.serviceId)
				return
			}
		}

	}
}

func TTLMonitorLogic(cc *CommonClient, state *MonitorState) error {
	// 执行健康检查动作
	err := cc.apiClient.Agent().UpdateTTL(cc.serviceId, "", "passing")
	if err == nil {
		logx.Infof("Service %s TTL updated successfully", cc.serviceId)
		state.RetryCount = 0
		state.BackoffTime = 1 * time.Second
		state.Ticker.Reset(state.OriginalTTL)
		return nil
	}

	registered, err := cc.registerServiceHealthStatus(api.HealthPassing)
	if err != nil {
		logx.Error(err)
	} else {
		registered = true
	}

	state.Mutex.Lock()
	defer state.Mutex.Unlock()
	if !registered && state.RetryCount < state.MaxRetries {
		logx.Infof("Attempting to re-register service %s (retry %d/%d)...", cc.serviceId, state.RetryCount+1, state.MaxRetries)
		err = cc.registerServiceWithPassingHealth()
		if err != nil {
			logx.Errorf("Failed to re-register service %s: %v. Retrying in %v", cc.serviceId, err, state.BackoffTime)
			state.RetryCount++

			state.Ticker.Reset(state.BackoffTime)
			nextBackoffTime := state.BackoffTime * 2
			if nextBackoffTime > state.MaxBackoffTime {
				nextBackoffTime = state.MaxBackoffTime
			}
			state.BackoffTime = nextBackoffTime
			return err
		}

		logx.Infof("Service %s re-registered successfully", cc.serviceId)
		state.RetryCount = 0
		state.BackoffTime = 1 * time.Second
		state.Ticker.Reset(state.OriginalTTL)
		return nil
	}

	// 达到最大重试次数
	if !registered {
		logx.Errorf("Max retries reached for service %s. Resetting retry counter and backoff time.", cc.serviceId)
		state.RetryCount = 0
		state.BackoffTime = 1 * time.Second
		state.Ticker.Reset(state.BackoffTime)
	}
	return nil
}

func HttpCheckMonitorFunc() MonitorFunc {

	return func(cc *CommonClient, stopCh <-chan struct{}) {
		var ttlTicker time.Duration
		if cc.consulConf.TTL > 0 {
			ttlTicker = time.Duration(cc.consulConf.TTL-1) * time.Second
			if ttlTicker < time.Second {
				ttlTicker = time.Second
			}
		} else {
			ttlTicker = 1 * time.Second
		}

		state := &MonitorState{
			RetryCount:     0,
			BackoffTime:    1 * time.Second,
			MaxRetries:     5,
			MaxBackoffTime: 30 * time.Second,
			OriginalTTL:    ttlTicker,
			Ticker:         time.NewTicker(ttlTicker),
		}
		defer state.Close()

		for {
			select {
			case <-state.Ticker.C:
				err := HttpMonitorLogic(cc, state)
				if err != nil {
					logx.Errorf("Monitor function error for service %s: %v", cc.serviceId, err)
				}
			case <-stopCh:
				logx.Infof("Service monitor for %s stopped gracefully", cc.serviceId)
				return
			}
		}
	}
}

func HttpMonitorLogic(cc *CommonClient, state *MonitorState) error {
	registered, err := cc.registerServiceHealthStatus(api.HealthPassing)
	if err != nil {
		logx.Error(err)
	} else {
		logx.Infof("Service %s health check passed", cc.serviceId)
		registered = true
	}

	state.Mutex.Lock()
	defer state.Mutex.Unlock()
	if !registered && state.RetryCount < state.MaxRetries {
		logx.Infof("Attempting to re-register service %s (retry %d/%d)...", cc.serviceId, state.RetryCount+1, state.MaxRetries)
		err = cc.registerServiceWithPassingHealth()
		if err != nil {
			logx.Errorf("Failed to re-register service %s: %v. Retrying in %v", cc.serviceId, err, state.BackoffTime)
			state.RetryCount++

			state.Ticker.Reset(state.BackoffTime)
			nextBackoffTime := state.BackoffTime * 2
			if nextBackoffTime > state.MaxBackoffTime {
				nextBackoffTime = state.MaxBackoffTime
			}
			state.BackoffTime = nextBackoffTime
			return err
		}

		logx.Infof("Service %s re-registered successfully", cc.serviceId)
		state.RetryCount = 0
		state.BackoffTime = 1 * time.Second
		state.Ticker.Reset(state.OriginalTTL)
		return nil
	}

	// 达到最大重试次数
	if !registered {
		logx.Errorf("Max retries reached for service %s. Resetting retry counter and backoff time.", cc.serviceId)
		state.RetryCount = 0
		state.BackoffTime = 1 * time.Second
		state.Ticker.Reset(state.BackoffTime)
	}
	return nil
}

func figureOutListenOn(listenOn string) string {
	fields := strings.Split(listenOn, ":")
	if len(fields) == 0 {
		return listenOn
	}

	host := fields[0]
	if len(host) > 0 && host != allEths {
		return listenOn
	}

	ip := os.Getenv(envPodIP)
	if len(ip) == 0 {
		ip = netx.InternalIp()
	}
	if len(ip) == 0 {
		return listenOn
	}

	return strings.Join(append([]string{ip}, fields[1:]...), ":")
}

func WithMonitorFuncs(funcs ...MonitorFunc) ServiceOption {
	return func(cc *CommonClient) {
		cc.monitorMutex.Lock()
		defer cc.monitorMutex.Unlock()
		cc.monitorFuncs = append(cc.monitorFuncs, funcs...)
	}
}
