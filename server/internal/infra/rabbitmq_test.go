package infra

import (
	"os"
	"testing"
	"time"

	"github.com/gcc798/xianyu.bot/internal/config"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestOpenRabbitMQValidatesSettingsBeforeConnecting(t *testing.T) {
	_, err := OpenRabbitMQ(config.RabbitMQ{})
	require.EqualError(t, err, "RabbitMQ URL 不能为空")

	settings := rabbitMQTestSettings("amqp://guest:guest@127.0.0.1:5672/")
	settings.RetryMediumTTLSeconds = settings.RetryShortTTLSeconds
	_, err = OpenRabbitMQ(settings)
	require.EqualError(t, err, "RabbitMQ 重试 TTL 必须按短、中、长递增")
}

func TestRabbitMQTopologyConfirmAndManualAck(t *testing.T) {
	url := os.Getenv("TEST_RABBITMQ_URL")
	if url == "" {
		t.Skip("TEST_RABBITMQ_URL is required")
	}
	settings := rabbitMQTestSettings(url)
	broker, err := OpenRabbitMQ(settings)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, broker.Close()) })
	require.NoError(t, broker.DeclareRAGTopology(settings))

	broker.publishMu.Lock()
	_, err = broker.publisher.QueuePurge(RAGBuildQueue, false)
	broker.publishMu.Unlock()
	require.NoError(t, err)
	for _, queue := range []string{ragBuildRetryShortQueue, ragBuildRetryMediumQueue, ragBuildRetryLongQueue, ragBuildDeadQueue} {
		_, err := broker.publisher.QueueInspect(queue)
		require.NoError(t, err)
	}

	messageID := uuid.NewString()
	firstConsumer, err := broker.OpenBuildConsumer(t.Context(), "integration-first-"+messageID)
	require.NoError(t, err)
	require.NoError(t, broker.PublishTask(t.Context(), RabbitTaskMessage{
		ID: messageID, Type: "rag.document.build.v1", RoutingKey: RAGBuildRoutingKey,
		Body: []byte(`{"schemaVersion":1,"taskId":42}`),
	}))

	firstDelivery := receiveRabbitDelivery(t, firstConsumer.Deliveries)
	require.Equal(t, messageID, firstDelivery.MessageId)
	require.Equal(t, "rag.document.build.v1", firstDelivery.Type)
	require.False(t, firstDelivery.Redelivered)
	require.NoError(t, firstConsumer.Close())

	secondConsumer, err := broker.OpenBuildConsumer(t.Context(), "integration-second-"+messageID)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, secondConsumer.Close()) })
	redelivery := receiveRabbitDelivery(t, secondConsumer.Deliveries)
	require.Equal(t, messageID, redelivery.MessageId)
	require.True(t, redelivery.Redelivered)
	require.NoError(t, redelivery.Ack(false))
}

func receiveRabbitDelivery(t *testing.T, deliveries <-chan RabbitDelivery) RabbitDelivery {
	t.Helper()
	select {
	case delivery, open := <-deliveries:
		require.True(t, open)
		return delivery
	case <-time.After(5 * time.Second):
		t.Fatal("等待 RabbitMQ 消息超时")
		return RabbitDelivery{}
	}
}

func rabbitMQTestSettings(url string) config.RabbitMQ {
	return config.RabbitMQ{
		URL: url, Prefetch: 2, PublishTimeoutSeconds: 5,
		RetryShortTTLSeconds: 10, RetryMediumTTLSeconds: 60, RetryLongTTLSeconds: 300,
	}
}
