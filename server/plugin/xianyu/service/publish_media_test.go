package service

import (
	"bytes"
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUploadPublishImageContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		require.NoError(t, request.ParseMultipartForm(1<<20))
		file, header, err := request.FormFile("file")
		require.NoError(t, err)
		require.NoError(t, file.Close())
		require.Equal(t, "publish-image.png", header.Filename)
		require.Equal(t, "credential-cookie", request.Header.Get("Cookie"))
		_, _ = writer.Write([]byte(`{"success":true,"object":{"url":"https://img.alicdn.com/uploaded.png","pix":"20x10"}}`))
	}))
	t.Cleanup(server.Close)
	var content bytes.Buffer
	require.NoError(t, png.Encode(&content, image.NewRGBA(image.Rect(0, 0, 20, 10))))

	result, err := uploadPublishImageContent(t.Context(), server.Client(), server.URL, "credential-cookie", content.Bytes(), "source.png")
	require.NoError(t, err)
	require.Equal(t, "https://img.alicdn.com/uploaded.png", result["url"])
	require.Equal(t, 20, result["widthSize"])
	require.Equal(t, 10, result["heightSize"])
}

func TestAllowedPublishImageHost(t *testing.T) {
	require.True(t, allowedPublishImageHost("img.alicdn.com"))
	require.True(t, allowedPublishImageHost("stream-upload.goofish.com"))
	require.False(t, allowedPublishImageHost("127.0.0.1"))
	require.False(t, allowedPublishImageHost("alicdn.com.example.test"))
}
