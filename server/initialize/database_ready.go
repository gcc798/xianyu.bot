package initialize

import (
	"sync"

	"github.com/flipped-aurora/gin-vue-admin/server/service/system"
)

func init() {
	// 注入数据库就绪回调到 service 层
	system.SetDBReadyCallback(func() {
		GetDatabaseReadyNotifier().NotifyDBReady()
	})
}

var (
	databaseReadyNotifier     *DatabaseReadyNotifier
	databaseReadyNotifierOnce sync.Once
)

// DatabaseReadyNotifier 数据库就绪通知器
type DatabaseReadyNotifier struct {
	mu          sync.RWMutex
	subscribers []func()
	dbReady     bool
}

// GetDatabaseReadyNotifier 获取全局通知器实例（单例）
func GetDatabaseReadyNotifier() *DatabaseReadyNotifier {
	databaseReadyNotifierOnce.Do(func() {
		databaseReadyNotifier = &DatabaseReadyNotifier{
			subscribers: make([]func(), 0),
			dbReady:     false,
		}
	})
	return databaseReadyNotifier
}

// Subscribe 订阅数据库就绪事件
// 如果数据库已经就绪，立即执行回调
// 否则将回调加入订阅列表，等待通知
func (n *DatabaseReadyNotifier) Subscribe(callback func()) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.dbReady {
		// 数据库已就绪，立即执行
		callback()
	} else {
		// 数据库未就绪，加入订阅列表
		n.subscribers = append(n.subscribers, callback)
	}
}

// NotifyDBReady 通知数据库已就绪
// 会触发所有订阅的回调函数
func (n *DatabaseReadyNotifier) NotifyDBReady() {
	n.mu.Lock()

	if n.dbReady {
		// 已经通知过了，避免重复通知
		n.mu.Unlock()
		return
	}

	n.dbReady = true

	// 复制订阅列表，避免在执行回调时持有锁
	callbacks := make([]func(), len(n.subscribers))
	copy(callbacks, n.subscribers)

	// 清空订阅列表
	n.subscribers = nil

	// 释放锁后再执行回调，避免死锁
	n.mu.Unlock()

	// 执行所有订阅的回调
	for _, callback := range callbacks {
		callback()
	}
}

// IsDBReady 检查数据库是否已就绪
func (n *DatabaseReadyNotifier) IsDBReady() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.dbReady
}
