package api

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestParseLastEventID(t *testing.T) {
	id, err := parseLastEventID(" 42 ")
	require.NoError(t, err)
	require.EqualValues(t, 42, id)
	_, err = parseLastEventID("invalid")
	require.Error(t, err)
}

func TestWriteAgentSSE(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	require.NoError(t, writeAgentSSE(ctx, 7, "run.completed", []byte(`{"runId":9}`)))
	require.Equal(t, "id: 7\nevent: run.completed\ndata: {\"runId\":9}\n\n", recorder.Body.String())
}
