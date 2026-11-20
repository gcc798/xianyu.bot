package agent

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/gcc798/xianyu.bot/internal/model"
	"github.com/mark3labs/mcp-go/mcp"
)

// mcpHealthCheck 保存一个可被并发请求共享的探测结果。
type mcpHealthCheck struct {
	done chan struct{}
	err  error
}

// CheckHealth 立即执行或合并指定 MCP 的独立协议健康探测。
func (r *MCPRegistry) CheckHealth(ctx context.Context, serverCode string) (model.MCPServer, error) {
	var server model.MCPServer
	if err := r.db.WithContext(ctx).Where("code = ? AND enabled = ?", serverCode, true).First(&server).Error; err != nil {
		return model.MCPServer{}, err
	}
	r.healthMu.Lock()
	if r.healthChecks == nil {
		r.healthChecks = make(map[string]*mcpHealthCheck)
	}
	if current := r.healthChecks[server.Code]; current != nil {
		r.healthMu.Unlock()
		select {
		case <-ctx.Done():
			return model.MCPServer{}, ctx.Err()
		case <-current.done:
			if err := r.db.WithContext(ctx).First(&server, server.ID).Error; err != nil {
				return model.MCPServer{}, err
			}
			return server, current.err
		}
	}
	check := &mcpHealthCheck{done: make(chan struct{})}
	r.healthChecks[server.Code] = check
	r.healthMu.Unlock()

	check.err = r.probeHealth(ctx, server)
	r.healthMu.Lock()
	delete(r.healthChecks, server.Code)
	close(check.done)
	r.healthMu.Unlock()
	if err := r.db.WithContext(ctx).First(&server, server.ID).Error; err != nil {
		return model.MCPServer{}, err
	}
	return server, check.err
}

// probeHealth 使用短期客户端完成握手和 Ping，并只更新健康投影。
func (r *MCPRegistry) probeHealth(ctx context.Context, server model.MCPServer) error {
	startedAt := time.Now()
	newClient := r.healthClient
	if newClient == nil {
		newClient = func(server model.MCPServer) (mcpHealthClient, error) { return newMCPHTTPClient(server) }
	}
	client, err := newClient(server)
	if err == nil {
		defer client.Close()
		timeout := time.Duration(server.RequestTimeoutSeconds) * time.Second
		if timeout <= 0 {
			timeout = 15 * time.Second
		}
		probeContext, cancel := context.WithTimeout(ctx, min(timeout, 15*time.Second))
		defer cancel()
		if err = client.Start(probeContext); err == nil {
			_, err = client.Initialize(probeContext, mcp.InitializeRequest{Params: mcp.InitializeParams{
				ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
				ClientInfo:      mcp.Implementation{Name: "xy-agent-health", Version: "1.0.0"},
			}})
		}
		if err == nil {
			err = client.Ping(probeContext)
		}
	}
	status := "available"
	message := ""
	if err != nil {
		status = "unavailable"
		message = safeMCPError(server, err)
	}
	updateErr := r.db.WithContext(ctx).Model(&model.MCPServer{}).Where("id = ?", server.ID).Updates(map[string]any{
		"last_status": status, "last_checked_at": time.Now(),
		"last_check_duration_millis": time.Since(startedAt).Milliseconds(), "last_error": message,
	}).Error
	return errors.Join(err, updateErr)
}

// StartHealthChecks 启动带周期抖动的全量 MCP 健康检查。
func (r *MCPRegistry) StartHealthChecks(ctx context.Context, interval time.Duration, concurrency int) {
	if interval <= 0 {
		interval = time.Minute
	}
	if concurrency < 1 {
		concurrency = 8
	}
	go func() {
		_ = r.checkAllHealth(ctx, concurrency)
		for {
			timer := time.NewTimer(jitteredHealthInterval(interval))
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
				_ = r.checkAllHealth(ctx, concurrency)
			}
		}
	}()
}

// checkAllHealth 以有界并发探测所有启用服务。
func (r *MCPRegistry) checkAllHealth(ctx context.Context, concurrency int) error {
	servers := make([]model.MCPServer, 0)
	if err := r.db.WithContext(ctx).Where("enabled = ?", true).Find(&servers).Error; err != nil {
		return err
	}
	rand.Shuffle(len(servers), func(i, j int) { servers[i], servers[j] = servers[j], servers[i] })
	slots := make(chan struct{}, max(1, concurrency))
	errorsFound := make(chan error, len(servers))
	var workers sync.WaitGroup
	for _, server := range servers {
		workers.Go(func() {
			select {
			case slots <- struct{}{}:
				defer func() { <-slots }()
			case <-ctx.Done():
				errorsFound <- ctx.Err()
				return
			}
			if _, err := r.CheckHealth(ctx, server.Code); err != nil {
				errorsFound <- err
			}
		})
	}
	workers.Wait()
	close(errorsFound)
	joined := make([]error, 0, len(errorsFound))
	for err := range errorsFound {
		joined = append(joined, err)
	}
	return errors.Join(joined...)
}

// jitteredHealthInterval 为固定周期增加正负百分之十抖动。
func jitteredHealthInterval(interval time.Duration) time.Duration {
	delta := interval / 10
	if delta <= 0 {
		return interval
	}
	return interval - delta + time.Duration(rand.Int63n(int64(2*delta)+1))
}
