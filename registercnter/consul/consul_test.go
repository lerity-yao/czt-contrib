package consul

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
)

func TestConsulRegisterTTL(t *testing.T) {
	// Set shorter TTL to speed up testing
	conf := Conf{
		Host:      "127.0.0.1:8501",
		Key:       "test-service",
		TTL:       5,
		CheckType: "ttl",
	}

	// Create a WaitGroup to wait for service registration goroutine
	var wg sync.WaitGroup
	wg.Add(1)

	service := MustNewService("127.0.0.1:8000", conf)

	err := service.Register()
	assert.Nil(t, err, "Service registration failed: %v", err)

	serviceID := service.GetServiceID()

	// Start service registration
	go func() {
		defer wg.Done()
	}()

	// Wait for service registration to complete
	time.Sleep(2 * time.Second)

	// Create Consul client for verification and operations
	client, err := api.NewClient(&api.Config{Scheme: "http", Address: conf.Host})
	assert.Nil(t, err, "Failed to create Consul client: %v", err)

	// Verify service is successfully registered
	services, err := client.Agent().Services()
	assert.Nil(t, err, "Failed to get services list: %v", err)
	_, exists := services[serviceID]
	assert.True(t, exists, "Service not successfully registered to Consul")
	fmt.Printf("✅ Service %s registered successfully\n", serviceID)

	// Wait for multiple TTL updates to observe TTL update logs
	fmt.Printf("⏱️  Waiting for TTL updates... (please observe TTL update logs)\n")
	time.Sleep(12 * time.Second) // Wait more than 2 TTL cycles to ensure seeing multiple updates

	// Simulate service disconnection - manually deregister service from Consul
	fmt.Printf("🔄 Simulating service disconnection, manually deregistering service %s...\n", serviceID)
	err = client.Agent().ServiceDeregister(serviceID)
	assert.Nil(t, err, "Failed to manually deregister service: %v", err)

	// Wait for service to detect deregistration and re-register
	time.Sleep(100 * time.Second) // Wait longer than TTL

	// Verify service has been re-registered
	services, err = client.Agent().Services()
	assert.Nil(t, err, "Failed to get services list again: %v", err)
	_, exists = services[serviceID]
	assert.True(t, exists, "Service failed to auto re-register after disconnection")
	fmt.Printf("✅ Service %s successfully auto re-registered\n", serviceID)

	// Wait for TTL update after re-registration to verify normal operation
	fmt.Printf("⏱️  Waiting for TTL updates after re-registration... (please observe TTL update logs)\n")
	time.Sleep(6 * time.Second)

	// Cleanup: deregister service after test completes
	defer func() {
		err := client.Agent().ServiceDeregister(serviceID)
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
		wg.Wait()
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

//
//func TestConsulRegisterHTTP(t *testing.T) {
//	// 使用127.0.0.1确保本地可访问
//	serverAddr := "172.17.0.1:8888"
//
//	// 创建一个简单的HTTP处理器
//	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		if r.URL.Path == "/healthz" {
//			fmt.Printf("Received health check request from %s\n", r.RemoteAddr)
//			w.WriteHeader(http.StatusOK)
//			w.Write([]byte(`{"status":"healthy"}`))
//		} else {
//			w.WriteHeader(http.StatusNotFound)
//		}
//	})
//
//	// 创建一个可关闭的HTTP服务器
//	server := &http.Server{
//		Addr:    serverAddr,
//		Handler: handler,
//	}
//
//	// 启动HTTP服务器
//	go func() {
//		fmt.Printf("Starting HTTP server on %s...\n", serverAddr)
//		err := server.ListenAndServe()
//		if err != nil && err != http.ErrServerClosed {
//			fmt.Printf("HTTP server error: %v\n", err)
//		}
//	}()
//
//	// 等待一小段时间确保服务器启动
//	time.Sleep(500 * time.Millisecond)
//
//	// 快速验证服务器是否启动成功
//	conn, err := net.DialTimeout("tcp", serverAddr, 1*time.Second)
//	if err != nil {
//		t.Fatalf("Failed to start HTTP server: %v", err)
//	}
//	conn.Close()
//	fmt.Printf("✅ HTTP server at %s is ready\n", serverAddr)
//
//	// 确保测试结束时关闭服务器
//	defer func() {
//		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
//		defer cancel()
//		if err := server.Shutdown(ctx); err != nil {
//			fmt.Printf("Error shutting down HTTP server: %v\n", err)
//		} else {
//			fmt.Println("HTTP server shutdown completed")
//		}
//	}()
//
//	// 设置HTTP健康检查配置
//	serviceKey := "test-service-http"
//	conf := Conf{
//		Host:      "127.0.0.1:8500",
//		Key:       serviceKey,
//		CheckType: "http",
//		CheckHttp: CheckHttpConf{
//			Host:     fmt.Sprintf("http://%s/healthz", serverAddr),
//			Interval: 5,
//			Timeout:  1,
//		},
//	}
//
//	// 创建一个有超时的context，避免测试无限阻塞
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	// 创建通道用于接收注册结果
//	registrationDone := make(chan bool, 1)
//
//	// 启动服务注册
//	go func() {
//		err := RegisterService(serverAddr, conf)
//		if err != nil {
//			fmt.Printf("Service registration failed: %v\n", err)
//		} else {
//			fmt.Printf("Service registration started successfully\n")
//		}
//		registrationDone <- true
//	}()
//
//	// 等待注册开始或超时
//	select {
//	case <-registrationDone:
//		fmt.Println("Service registration goroutine started")
//	case <-ctx.Done():
//		t.Fatalf("Registration goroutine startup timeout")
//	}
//
//	// 等待服务注册完成
//	time.Sleep(20 * time.Second)
//
//	// 创建Consul客户端进行验证
//	client, err := api.NewClient(&api.Config{Scheme: "http", Address: conf.Host})
//	assert.Nil(t, err, "Failed to create Consul client: %v", err)
//
//	// 获取所有服务，查找与我们的serviceKey匹配的服务ID
//	services, err := client.Agent().Services()
//	assert.Nil(t, err, "Failed to get services list: %v", err)
//
//	var serviceID string
//	serviceFound := false
//	for id := range services {
//		if strings.Contains(id, serviceKey) {
//			serviceID = id
//			serviceFound = true
//			fmt.Printf("✅ Found service with ID: %s\n", serviceID)
//			break
//		}
//	}
//
//	assert.True(t, serviceFound, "Service not successfully registered to Consul")
//
//	// 获取并验证健康检查配置
//	checks, err := client.Agent().Checks()
//	assert.Nil(t, err, "Failed to get checks list: %v", err)
//
//	checkFound := false
//	for _, check := range checks {
//		if strings.Contains(check.CheckID, serviceKey) && check.Type == "http" {
//			checkFound = true
//			fmt.Printf("✅ Found HTTP health check for service %s\n", serviceKey)
//			break
//		}
//	}
//	assert.True(t, checkFound, "HTTP health check not found for service")
//
//	// 模拟服务断开 - 手动从Consul注销服务
//	if serviceID != "" {
//		fmt.Printf("🔄 Simulating service disconnection, manually deregistering service %s...\n", serviceID)
//		err = client.Agent().ServiceDeregister(serviceID)
//		assert.Nil(t, err, "Failed to manually deregister service: %v", err)
//
//		// 等待服务检测到注销并重新注册
//		time.Sleep(10 * time.Second)
//
//		// 验证服务已重新注册
//		services, err = client.Agent().Services()
//		assert.Nil(t, err, "Failed to get services list again: %v", err)
//
//		serviceReRegistered := false
//		for id := range services {
//			if strings.Contains(id, serviceKey) {
//				serviceReRegistered = true
//				serviceID = id // 更新serviceID，以防重新注册时生成了新的ID
//				fmt.Printf("✅ Service %s successfully auto re-registered\n", serviceID)
//				break
//			}
//		}
//		assert.True(t, serviceReRegistered, "Service failed to auto re-register after disconnection")
//	}
//
//	// 清理：测试完成后注销服务
//	defer func() {
//		if serviceID != "" {
//			err := client.Agent().ServiceDeregister(serviceID)
//			if err != nil {
//				fmt.Printf("Failed to cleanup service: %v\n", err)
//			} else {
//				fmt.Printf("🧹 Test completed, service %s cleaned up\n", serviceID)
//			}
//		}
//	}()
//
//	fmt.Println("✅ Test completed successfully")
//}
