package infra

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gcc798/xianyu.bot/internal/config"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	// RAGTaskExchange 是全部 RAG 异步任务进入的持久 Topic Exchange。
	RAGTaskExchange = "xianyu.rag.tasks"
	// RAGRetryExchange 将失败消息路由到不同 TTL 的重试队列。
	RAGRetryExchange = "xianyu.rag.retry"
	// RAGDeadExchange 收集明确不可重试或超过重试上限的消息。
	RAGDeadExchange = "xianyu.rag.dead"

	// RAGBuildQueue 是文档构建 Worker 的主队列。
	RAGBuildQueue = "xianyu.rag.build"
	// RAGBuildRoutingKey 是文档构建事件的稳定路由键。
	RAGBuildRoutingKey = "rag.build"
	// RAGBuildDeadRoutingKey 是构建任务死信的路由键。
	RAGBuildDeadRoutingKey = "rag.build.dead"

	// RAGRetryShortRoutingKey 选择短退避队列。
	RAGRetryShortRoutingKey = "rag.build.retry.short"
	// RAGRetryMediumRoutingKey 选择中退避队列。
	RAGRetryMediumRoutingKey = "rag.build.retry.medium"
	// RAGRetryLongRoutingKey 选择长退避队列。
	RAGRetryLongRoutingKey = "rag.build.retry.long"
)

const (
	ragBuildRetryShortQueue  = "xianyu.rag.build.retry.short"
	ragBuildRetryMediumQueue = "xianyu.rag.build.retry.medium"
	ragBuildRetryLongQueue   = "xianyu.rag.build.retry.long"
	ragBuildDeadQueue        = "xianyu.rag.build.dead"
)

// RabbitMQ 持有发布专用连接和启用 Confirm 的 Channel。
// 消费者使用独立连接，避免 Broker 流控同时阻塞消息确认和消费 Ack。
type RabbitMQ struct {
	url            string           // url 仅用于创建独立消费连接，不得写入日志。
	prefetch       int              // prefetch 限制单个 Consumer 的未确认消息数。
	publishTimeout time.Duration    // publishTimeout 限制等待 Broker Confirm 的时间。
	connection     *amqp.Connection // connection 是发布专用连接。
	publisher      *amqp.Channel    // publisher 是启用 Confirm 的发布 Channel。
	publishMu      sync.Mutex       // publishMu 串行化 Channel 声明和带 Confirm 的发布。
	state          atomic.Uint32    // state 保存发布连接的最新生命周期状态，供就绪探针读取。
	fatalErrors    chan error       // fatalErrors 通知调用方自动恢复已耗尽，需要重启进程。
}

// RabbitTaskMessage 是发布到任务 Exchange 的持久消息。
type RabbitTaskMessage struct {
	ID         string // ID 是 Outbox ID 等稳定幂等标识。
	Type       string // Type 是带版本的事件类型。
	RoutingKey string // RoutingKey 选择主队列或某档重试队列。
	Body       []byte // Body 只包含任务定位字段，不得承载凭据或完整原文。
}

// RabbitDelivery 是需要由 Worker 显式 Ack 或 Reject 的 AMQP 投递。
type RabbitDelivery = amqp.Delivery

// RabbitConsumer 是使用手动 Ack 的独立 AMQP Consumer。
type RabbitConsumer struct {
	Deliveries <-chan RabbitDelivery // Deliveries 必须持续读取；成功后 Ack，失败时 Reject 或重试发布。
	connection *amqp.Connection      // connection 与发布连接隔离。
	channel    *amqp.Channel         // channel 持有 Consumer 和未确认投递。
}

// OpenRabbitMQ 建立发布连接并启用 Publisher Confirm。
func OpenRabbitMQ(settings config.RabbitMQ) (*RabbitMQ, error) {
	if err := validateRabbitMQSettings(settings); err != nil {
		return nil, err
	}
	properties := amqp.NewConnectionProperties()
	properties.SetClientConnectionName("xianyu-rag-worker-publisher")
	connection, err := amqp.DialConfig(settings.URL, amqp.Config{
		Properties: properties,
		Recovery:   new(amqp.Recovery),
	})
	if err != nil {
		return nil, fmt.Errorf("连接 RabbitMQ: %w", err)
	}
	publisher, err := connection.Channel()
	if err != nil {
		_ = connection.Close()
		return nil, fmt.Errorf("创建 RabbitMQ 发布 Channel: %w", err)
	}
	if err := publisher.Confirm(false); err != nil {
		_ = publisher.Close()
		_ = connection.Close()
		return nil, fmt.Errorf("启用 RabbitMQ Publisher Confirm: %w", err)
	}
	broker := &RabbitMQ{
		url: settings.URL, prefetch: settings.Prefetch,
		publishTimeout: time.Duration(settings.PublishTimeoutSeconds) * time.Second,
		connection:     connection, publisher: publisher,
		fatalErrors: make(chan error, 1),
	}
	broker.trackConnectionState()
	return broker, nil
}

// DeclareRAGTopology 幂等声明构建主队列、三级 TTL 重试队列和死信队列。
func (b *RabbitMQ) DeclareRAGTopology(settings config.RabbitMQ) error {
	if b == nil || b.publisher == nil {
		return errors.New("RabbitMQ 发布 Channel 未初始化")
	}
	if err := validateRetryTTL(settings); err != nil {
		return err
	}

	b.publishMu.Lock()
	defer b.publishMu.Unlock()

	for _, exchange := range []struct {
		name string
		kind string
	}{
		{name: RAGTaskExchange, kind: "topic"},
		{name: RAGRetryExchange, kind: "direct"},
		{name: RAGDeadExchange, kind: "direct"},
	} {
		if err := b.publisher.ExchangeDeclare(exchange.name, exchange.kind, true, false, false, false, nil); err != nil {
			return fmt.Errorf("声明 RabbitMQ Exchange %s: %w", exchange.name, err)
		}
	}

	mainArgs := amqp.Table{
		"x-dead-letter-exchange":    RAGDeadExchange,
		"x-dead-letter-routing-key": RAGBuildDeadRoutingKey,
	}
	if _, err := b.publisher.QueueDeclare(RAGBuildQueue, true, false, false, false, mainArgs); err != nil {
		return fmt.Errorf("声明 RAG 构建主队列: %w", err)
	}
	if err := b.publisher.QueueBind(RAGBuildQueue, RAGBuildRoutingKey, RAGTaskExchange, false, nil); err != nil {
		return fmt.Errorf("绑定 RAG 构建主队列: %w", err)
	}

	retries := []struct {
		queue      string
		routingKey string
		ttlSeconds int
	}{
		{queue: ragBuildRetryShortQueue, routingKey: RAGRetryShortRoutingKey, ttlSeconds: settings.RetryShortTTLSeconds},
		{queue: ragBuildRetryMediumQueue, routingKey: RAGRetryMediumRoutingKey, ttlSeconds: settings.RetryMediumTTLSeconds},
		{queue: ragBuildRetryLongQueue, routingKey: RAGRetryLongRoutingKey, ttlSeconds: settings.RetryLongTTLSeconds},
	}
	for _, retry := range retries {
		args := amqp.Table{
			"x-message-ttl":             int64(retry.ttlSeconds) * int64(time.Second/time.Millisecond),
			"x-dead-letter-exchange":    RAGTaskExchange,
			"x-dead-letter-routing-key": RAGBuildRoutingKey,
		}
		if _, err := b.publisher.QueueDeclare(retry.queue, true, false, false, false, args); err != nil {
			return fmt.Errorf("声明 RAG 重试队列 %s: %w", retry.queue, err)
		}
		if err := b.publisher.QueueBind(retry.queue, retry.routingKey, RAGRetryExchange, false, nil); err != nil {
			return fmt.Errorf("绑定 RAG 重试队列 %s: %w", retry.queue, err)
		}
	}

	if _, err := b.publisher.QueueDeclare(ragBuildDeadQueue, true, false, false, false, nil); err != nil {
		return fmt.Errorf("声明 RAG 构建死信队列: %w", err)
	}
	if err := b.publisher.QueueBind(ragBuildDeadQueue, RAGBuildDeadRoutingKey, RAGDeadExchange, false, nil); err != nil {
		return fmt.Errorf("绑定 RAG 构建死信队列: %w", err)
	}
	return nil
}

// PublishTask 发布持久任务并等待 Broker Confirm；未确认时调用方必须保留 Outbox 待重试状态。
func (b *RabbitMQ) PublishTask(ctx context.Context, message RabbitTaskMessage) error {
	if b == nil || b.publisher == nil {
		return errors.New("RabbitMQ 发布 Channel 未初始化")
	}
	if strings.TrimSpace(message.ID) == "" || strings.TrimSpace(message.Type) == "" || strings.TrimSpace(message.RoutingKey) == "" {
		return errors.New("RabbitMQ 任务消息缺少 ID、类型或路由键")
	}
	if !slices.Contains([]string{RAGBuildRoutingKey, RAGRetryShortRoutingKey, RAGRetryMediumRoutingKey, RAGRetryLongRoutingKey, RAGBuildDeadRoutingKey}, message.RoutingKey) {
		return fmt.Errorf("RabbitMQ 不支持任务路由键: %s", message.RoutingKey)
	}
	if len(message.Body) == 0 {
		return errors.New("RabbitMQ 任务消息正文不能为空")
	}

	b.publishMu.Lock()
	defer b.publishMu.Unlock()

	publishCtx, cancel := context.WithTimeout(ctx, b.publishTimeout)
	defer cancel()
	exchange := RAGTaskExchange
	if strings.HasPrefix(message.RoutingKey, "rag.build.retry.") {
		exchange = RAGRetryExchange
	} else if message.RoutingKey == RAGBuildDeadRoutingKey {
		exchange = RAGDeadExchange
	}
	confirmation, err := b.publisher.PublishWithDeferredConfirmWithContext(
		publishCtx,
		exchange,
		message.RoutingKey,
		true,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			MessageId:    message.ID,
			Type:         message.Type,
			Timestamp:    time.Now().UTC(),
			Body:         message.Body,
		},
	)
	if err != nil {
		return fmt.Errorf("发布 RabbitMQ 任务消息: %w", err)
	}
	if confirmation == nil {
		return errors.New("RabbitMQ 未返回 Publisher Confirm 句柄")
	}
	acknowledged, err := confirmation.WaitContext(publishCtx)
	if err != nil {
		return fmt.Errorf("等待 RabbitMQ Publisher Confirm: %w", err)
	}
	if !acknowledged {
		return errors.New("RabbitMQ 拒绝任务消息")
	}
	return nil
}

// OpenBuildConsumer 创建使用独立连接、固定 Prefetch 和手动 Ack 的构建任务 Consumer。
func (b *RabbitMQ) OpenBuildConsumer(ctx context.Context, consumerName string) (*RabbitConsumer, error) {
	if b == nil || strings.TrimSpace(b.url) == "" {
		return nil, errors.New("RabbitMQ 消费连接配置为空")
	}
	properties := amqp.NewConnectionProperties()
	properties.SetClientConnectionName("xianyu-rag-worker-consumer")
	connection, err := amqp.DialConfig(b.url, amqp.Config{
		Properties: properties,
		Recovery:   new(amqp.Recovery),
	})
	if err != nil {
		return nil, fmt.Errorf("连接 RabbitMQ Consumer: %w", err)
	}
	channel, err := connection.Channel()
	if err != nil {
		_ = connection.Close()
		return nil, fmt.Errorf("创建 RabbitMQ Consumer Channel: %w", err)
	}
	if err := channel.Qos(b.prefetch, 0, false); err != nil {
		_ = channel.Close()
		_ = connection.Close()
		return nil, fmt.Errorf("设置 RabbitMQ Consumer Prefetch: %w", err)
	}
	deliveries, err := channel.ConsumeWithContext(ctx, RAGBuildQueue, consumerName, false, false, false, false, nil)
	if err != nil {
		_ = channel.Close()
		_ = connection.Close()
		return nil, fmt.Errorf("启动 RabbitMQ 构建任务 Consumer: %w", err)
	}
	return &RabbitConsumer{Deliveries: deliveries, connection: connection, channel: channel}, nil
}

// Health 检查发布连接当前是否仍可用；连接恢复期间返回未就绪。
func (b *RabbitMQ) Health(context.Context) error {
	if b == nil || b.connection == nil || b.connection.IsClosed() || amqp.LifeCycleState(b.state.Load()) != amqp.StateOpen {
		return errors.New("RabbitMQ 连接不可用")
	}
	return nil
}

// FatalErrors 返回自动恢复耗尽等不可自愈连接错误；调用方应退出并交给进程管理器重启。
func (b *RabbitMQ) FatalErrors() <-chan error {
	return b.fatalErrors
}

func (b *RabbitMQ) trackConnectionState() {
	b.state.Store(uint32(amqp.StateOpen))
	stateChanges := make(chan *amqp.StateChanged, 1)
	b.connection.NotifyStateChange(stateChanges)
	go func() {
		for change := range stateChanges {
			b.state.Store(uint32(change.To))
			if change.To == amqp.StateClosed && change.Err != nil {
				select {
				case b.fatalErrors <- change.Err:
				default:
				}
			}
		}
	}()
}

// Close 关闭发布 Channel 和连接。
func (b *RabbitMQ) Close() error {
	if b == nil {
		return nil
	}
	var closeErrors []error
	if b.publisher != nil && !b.publisher.IsClosed() {
		closeErrors = append(closeErrors, b.publisher.Close())
	}
	if b.connection != nil && !b.connection.IsClosed() {
		closeErrors = append(closeErrors, b.connection.Close())
	}
	return errors.Join(closeErrors...)
}

// Close 取消 Consumer 并关闭其独立连接；未 Ack 的消息由 RabbitMQ 自动重新入队。
func (c *RabbitConsumer) Close() error {
	if c == nil {
		return nil
	}
	var closeErrors []error
	if c.channel != nil && !c.channel.IsClosed() {
		closeErrors = append(closeErrors, c.channel.Close())
	}
	if c.connection != nil && !c.connection.IsClosed() {
		closeErrors = append(closeErrors, c.connection.Close())
	}
	return errors.Join(closeErrors...)
}

func validateRabbitMQSettings(settings config.RabbitMQ) error {
	if strings.TrimSpace(settings.URL) == "" {
		return errors.New("RabbitMQ URL 不能为空")
	}
	if settings.Prefetch <= 0 {
		return errors.New("RabbitMQ Prefetch 必须大于零")
	}
	if settings.PublishTimeoutSeconds <= 0 {
		return errors.New("RabbitMQ 发布确认超时必须大于零")
	}
	return validateRetryTTL(settings)
}

func validateRetryTTL(settings config.RabbitMQ) error {
	if settings.RetryShortTTLSeconds <= 0 || settings.RetryMediumTTLSeconds <= 0 || settings.RetryLongTTLSeconds <= 0 {
		return errors.New("RabbitMQ 重试 TTL 必须大于零")
	}
	if settings.RetryShortTTLSeconds >= settings.RetryMediumTTLSeconds || settings.RetryMediumTTLSeconds >= settings.RetryLongTTLSeconds {
		return errors.New("RabbitMQ 重试 TTL 必须按短、中、长递增")
	}
	return nil
}
