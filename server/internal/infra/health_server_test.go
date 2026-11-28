package infra

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHealthHandlerSeparatesLivenessAndReadiness(t *testing.T) {
	readinessErr := errors.New("RabbitMQ unavailable")
	handler := healthHandler(func(context.Context) error { return readinessErr })

	liveness := httptest.NewRecorder()
	handler.ServeHTTP(liveness, httptest.NewRequest(http.MethodGet, "/health", nil))
	require.Equal(t, http.StatusOK, liveness.Code)
	require.Equal(t, "ok", liveness.Body.String())

	notReady := httptest.NewRecorder()
	handler.ServeHTTP(notReady, httptest.NewRequest(http.MethodGet, "/ready", nil))
	require.Equal(t, http.StatusServiceUnavailable, notReady.Code)
	require.NotContains(t, notReady.Body.String(), readinessErr.Error())
}

func TestHealthHandlerReportsReady(t *testing.T) {
	handler := healthHandler(func(context.Context) error { return nil })
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/ready", nil))

	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, "ready", response.Body.String())
}
