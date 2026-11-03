package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMTopSellerRequestCarriesFishShopContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "COMMONPRO", r.URL.Query().Get("idle_site_biz_code"))
		require.Equal(t, "a21107h.42826273.0.0", r.URL.Query().Get("spm_cnt"))
		require.Equal(t, "COMMONPRO", r.Header.Get("idle_site_biz_code"))
		_, _ = w.Write([]byte(`{"ret":["SUCCESS::调用成功"],"data":{"itemId":"1"}}`))
	}))
	t.Cleanup(server.Close)

	client := &MTopClient{client: server.Client(), endpoint: server.URL}
	envelope, _, err := client.call(t.Context(), "_m_h5_tk=token_1", "mtop.idle.pc.backend.idleitem.editdetail", map[string]any{"itemId": "1"}, true)
	require.NoError(t, err)
	require.JSONEq(t, `{"itemId":"1"}`, string(envelope.Data))
}

func TestMTopSupportsCategoryRecommendationVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/mtop.taobao.idle.kgraph.pc.property.recommend/2.0/", r.URL.Path)
		require.Equal(t, "2.0", r.URL.Query().Get("v"))
		require.Equal(t, "mtop.taobao.idle.kgraph.pc.property.recommend", r.URL.Query().Get("api"))
		_, _ = w.Write([]byte(`{"ret":["SUCCESS::调用成功"],"data":{}}`))
	}))
	t.Cleanup(server.Close)

	client := &MTopClient{client: server.Client(), endpoint: server.URL}
	_, _, err := client.call(t.Context(), "_m_h5_tk=token_1", "mtop.taobao.idle.kgraph.pc.property.recommend/2.0", map[string]any{}, true)
	require.NoError(t, err)
}
