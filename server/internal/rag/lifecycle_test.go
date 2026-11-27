package rag

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCanTransitionIndexVersion(t *testing.T) {
	tests := []struct {
		name string            // name 是测试场景。
		from IndexVersionState // from 是当前索引状态。
		to   IndexVersionState // to 是目标索引状态。
		want bool              // want 是迁移是否允许。
	}{
		{name: "publish staging", from: IndexStaging, to: IndexPublished, want: true},
		{name: "fail staging", from: IndexStaging, to: IndexFailed, want: true},
		{name: "retire published", from: IndexPublished, to: IndexRetired, want: true},
		{name: "rollback retired", from: IndexRetired, to: IndexPublished, want: true},
		{name: "failed is immutable", from: IndexFailed, to: IndexStaging, want: false},
		{name: "cannot publish twice", from: IndexPublished, to: IndexPublished, want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.want, CanTransitionIndexVersion(test.from, test.to))
		})
	}
}

func TestCanTransitionBuildTask(t *testing.T) {
	tests := []struct {
		name string         // name 是测试场景。
		from BuildTaskState // from 是当前任务状态。
		to   BuildTaskState // to 是目标任务状态。
		want bool           // want 是迁移是否允许。
	}{
		{name: "claim queued", from: BuildQueued, to: BuildRunning, want: true},
		{name: "cancel queued", from: BuildQueued, to: BuildCancelled, want: true},
		{name: "recover expired lease", from: BuildRunning, to: BuildQueued, want: true},
		{name: "finish running", from: BuildRunning, to: BuildSucceeded, want: true},
		{name: "fail running", from: BuildRunning, to: BuildFailed, want: true},
		{name: "retry failed", from: BuildFailed, to: BuildQueued, want: true},
		{name: "succeeded is terminal", from: BuildSucceeded, to: BuildQueued, want: false},
		{name: "cancelled is terminal", from: BuildCancelled, to: BuildQueued, want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.want, CanTransitionBuildTask(test.from, test.to))
		})
	}
}
