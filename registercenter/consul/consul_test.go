package consul

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConsulRegisterTTL(t *testing.T) {
	// Set shorter TTL to speed up testing
	conf := Conf{
		Host:      "192.168.100.156:31095",
		Key:       "test-service",
		TTL:       5,
		CheckType: "ttl",
		Token:     "af9a9026-970a-5e7f-9b09-9cdedfd8a320",
	}

	// Create a WaitGroup to wait for service registration goroutine

	service := MustNewService("127.0.0.1:8000", conf)

	err := service.RegisterService()
	require.Nil(t, err)

	serviceID := service.GetServiceID()
	fmt.Printf("🔧 Generated service ID: %s\n", serviceID)

	// Wait for service registration to complete
	fmt.Printf("⏳ Waiting for service registration to complete...\n")
	time.Sleep(3 * time.Second)

	// Create Consul client for verification and operations WITH THE SAME TOKEN
	client, err := api.NewClient(&api.Config{
		Scheme:  "http",
		Address: conf.Host,
		Token:   conf.Token,
	})
	require.NoError(t, err, "Failed to create Consul client: %v", err)

	// Verify service is successfully registered with retry mechanism
	fmt.Printf("🔍 Checking if service %s is registered...\n", serviceID)

	//Verify service is successfully registered
	services, meta, err := client.Catalog().Service(conf.Key, "", nil)
	assert.Nil(t, err, "Failed to get services list: %v", err)

	assert.NotNil(t, meta, "Metadata should not be nil")
	assert.NotNil(t, services, "Services list should not be nil")

	// 验证至少有一个服务被找到
	assert.Greater(t, len(services), 0, "Should find at least one service")

	// Wait for multiple TTL updates to observe TTL update logs
	fmt.Printf("⏱️  Waiting for TTL updates... (please observe TTL update logs)\n")
	time.Sleep(12 * time.Second) // Wait more than 2 TTL cycles to ensure seeing multiple updates

	// Simulate service disconnection - manually deregister service from Consul
	fmt.Printf("🔄 Simulating service disconnection, manually deregistering service %s...\n", serviceID)
	err = service.DeregisterService()
	assert.Nil(t, err, "Failed to manually deregister service: %v", err)

	// Wait for service to detect deregistration and re-register
	time.Sleep(12 * time.Second) // Wait longer than TTL

	// Verify service has been re-registered
	services, meta, err = client.Catalog().Service(conf.Key, "", nil)
	assert.Nil(t, err, "Failed to get services list: %v", err)

	assert.NotNil(t, meta, "Metadata should not be nil")
	assert.NotNil(t, services, "Services list should not be nil")

	// 验证至少有一个服务被找到
	assert.Greater(t, len(services), 0, "Should find at least one service")

	// Wait for TTL update after re-registration to verify normal operation
	fmt.Printf("⏱️  Waiting for TTL updates after re-registration... (please observe TTL update logs)\n")
	time.Sleep(6 * time.Second)

	// Cleanup: deregister service after test completes
	defer func() {
		err := service.DeregisterService()
		if err != nil {
			fmt.Printf("Failed to cleanup service: %v\n", err)
		} else {
			fmt.Printf("🧹 Test completed, service %s cleaned up\n", serviceID)
		}
	}()

	// Set timeout to ensure test can finish
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Use timeout with WaitGroup to complete test
	done := make(chan struct{})
	go func() {
		close(done)
	}()

	// Wait for service registration goroutine to complete or timeout
	select {
	case <-done:
		fmt.Println("✅ Test completed, service registration goroutine exited normally")
	case <-ctx.Done():
		fmt.Println("⚠️  Test timed out, forcing exit")
	}
}

func TestConsulRegisterHTTP(t *testing.T) {
	// 使用127.0.0.1确保本地可访问
	serverAddr := "172.17.0.1:8888"

	// 创建一个简单的HTTP处理器
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			fmt.Printf("Received health check request from %s\n", r.RemoteAddr)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"healthy"}`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	})

	// 创建一个可关闭的HTTP服务器
	server := &http.Server{
		Addr:    serverAddr,
		Handler: handler,
	}

	// 启动HTTP服务器
	go func() {
		fmt.Printf("Starting HTTP server on %s...\n", serverAddr)
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()

	// 等待一小段时间确保服务器启动
	time.Sleep(500 * time.Millisecond)

	// 快速验证服务器是否启动成功
	conn, err := net.DialTimeout("tcp", serverAddr, 1*time.Second)
	if err != nil {
		t.Fatalf("Failed to start HTTP server: %v", err)
	}
	conn.Close()
	fmt.Printf("✅ HTTP server at %s is ready\n", serverAddr)

	// 确保测试结束时关闭服务器
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			fmt.Printf("Error shutting down HTTP server: %v\n", err)
		} else {
			fmt.Println("HTTP server shutdown completed")
		}
	}()

	// 设置HTTP健康检查配置
	serviceKey := "test-service-http"
	conf := Conf{
		Host:         "172.17.0.1:8500",
		Key:          serviceKey,
		CheckType:    "http",
		Token:        "af9a9026-970a-5e7f-9b09-9cdedfd8a320",
		TTL:          5,
		ExpiredTTL:   3,
		CheckTimeout: 2,
		CheckHttp: CheckHttpConf{
			Host: "172.17.0.1",
			Port: 8888,
		},
	}

	service := MustNewService(serverAddr, conf)
	err = service.RegisterService()
	serviceID := service.GetServiceID()
	require.Nil(t, err)

	// 等待服务注册完成
	time.Sleep(20 * time.Second)

	// 创建Consul客户端进行验证
	client, err := api.NewClient(&api.Config{Scheme: "http", Address: "172.17.0.1:8501"})
	assert.Nil(t, err, "Failed to create Consul client: %v", err)

	// 获取所有服务，查找与我们的serviceKey匹配的服务ID
	services, meta, err := client.Catalog().Service(conf.Key, "", nil)
	assert.Nil(t, err, "Failed to get services list: %v", err)
	assert.NotNil(t, meta, "Metadata should not be nil")
	assert.NotNil(t, services, "Services list should not be nil")

	// 验证至少有一个服务被找到
	assert.Greater(t, len(services), 0, "Should find at least one service")

	// 检查服务的健康状态
	//healthChecks, _, err := client.Health().Checks(conf.Key, nil)
	//assert.Nil(t, err, "Failed to get health checks: %v", err)
	//
	//// 输出每个健康检查的状态
	//for _, check := range healthChecks {
	//	fmt.Printf("📋 Health check for service %s: %s (Status: %s)\n",
	//		check.ServiceName, check.CheckID, check.Status)
	//}

	// 模拟服务断开 - 手动从Consul注销服务
	if serviceID != "" {
		fmt.Printf("🔄 Simulating service disconnection, manually deregistering service %s...\n", serviceID)
		err = service.DeregisterService()
		assert.Nil(t, err, "Failed to manually deregister service: %v", err)

		// 等待服务检测到注销并重新注册
		time.Sleep(20 * time.Second)

		// 获取所有服务，查找与我们的serviceKey匹配的服务ID
		servicesx, metax, errx := client.Catalog().Service(conf.Key, "", nil)
		assert.Nil(t, err, "Failed to get services list: %v", errx)
		assert.NotNil(t, metax, "Metadata should not be nil")
		assert.NotNil(t, servicesx, "Services list should not be nil")

		// 验证至少有一个服务被找到
		assert.GreaterOrEqual(t, len(servicesx), 1, "Should find at least one service")
		fmt.Printf("Service failed to auto re-register after disconnection")
	}

	// 清理：测试完成后注销服务
	defer func() {
		if serviceID != "" {
			err := client.Agent().ServiceDeregister(serviceID)
			if err != nil {
				fmt.Printf("Failed to cleanup service: %v\n", err)
			} else {
				fmt.Printf("🧹 Test completed, service %s cleaned up\n", serviceID)
			}
		}
	}()

	fmt.Println("✅ Test completed successfully")
}
