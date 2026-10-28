package system

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCaptchaRequired(t *testing.T) {
	tests := []struct {
		name     string
		limit    int
		attempts int
		want     bool
	}{
		{name: "disabled", limit: -1, attempts: 100, want: false},
		{name: "always enabled", limit: 0, attempts: 1, want: true},
		{name: "below threshold", limit: 3, attempts: 3, want: false},
		{name: "above threshold", limit: 3, attempts: 4, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, captchaRequired(tt.limit, tt.attempts))
		})
	}
}
