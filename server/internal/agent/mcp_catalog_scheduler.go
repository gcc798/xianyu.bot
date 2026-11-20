package agent

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/gcc798/xianyu.bot/internal/model"
)

// StartCatalogRefresh 加载持久快照并启动目录 TTL 扫描。
func (r *MCPRegistry) StartCatalogRefresh(ctx context.Context, scanInterval time.Duration, concurrency int) error {
	if err := r.LoadCatalog(ctx); err != nil {
		return err
	}
	if scanInterval <= 0 {
		scanInterval = 30 * time.Second
	}
	if concurrency < 1 {
		concurrency = 4
	}
	go func() {
		_ = r.refreshDueCatalogs(ctx, concurrency)
		ticker := time.NewTicker(scanInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_ = r.refreshDueCatalogs(ctx, concurrency)
			}
		}
	}()
	return nil
}

// refreshDueCatalogs 使用固定并发刷新首次发现、stale 或超过 TTL 的 MCP。
func (r *MCPRegistry) refreshDueCatalogs(ctx context.Context, concurrency int) error {
	if !r.catalogRefreshMu.TryLock() {
		return nil
	}
	defer r.catalogRefreshMu.Unlock()
	servers := make([]model.MCPServer, 0)
	if err := r.db.WithContext(ctx).Where("enabled = ?", true).Order("id").Find(&servers).Error; err != nil {
		return err
	}
	due := make([]model.MCPServer, 0, len(servers))
	now := time.Now()
	for _, server := range servers {
		ttl := time.Duration(server.CatalogTTLSeconds) * time.Second
		if ttl <= 0 {
			ttl = time.Hour
		}
		if server.CatalogStatus != "available" || server.CatalogRefreshedAt == nil || now.Sub(*server.CatalogRefreshedAt) >= ttl {
			due = append(due, server)
		}
	}
	if len(due) == 0 {
		return nil
	}
	concurrency = max(1, concurrency)
	slots := make(chan struct{}, concurrency)
	errorsFound := make(chan error, len(due))
	var workers sync.WaitGroup
	for _, server := range due {
		workers.Go(func() {
			select {
			case slots <- struct{}{}:
				defer func() { <-slots }()
			case <-ctx.Done():
				errorsFound <- ctx.Err()
				return
			}
			if err := r.RefreshCatalog(ctx, server.Code); err != nil {
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
