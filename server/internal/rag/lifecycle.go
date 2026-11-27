package rag

// IndexVersionState 标识不可变索引版本的发布生命周期。
type IndexVersionState string

const (
	// IndexStaging 表示索引仍在构建或验证，在线查询不可见。
	IndexStaging IndexVersionState = "staging"
	// IndexPublished 表示索引已经过验证并可用于在线查询。
	IndexPublished IndexVersionState = "published"
	// IndexRetired 表示索引已被新版本替换，但仍可在保留期内回滚。
	IndexRetired IndexVersionState = "retired"
	// IndexFailed 表示索引构建或验证失败，不得发布。
	IndexFailed IndexVersionState = "failed"
)

// CanTransitionIndexVersion 判断索引版本能否在不修改内容的前提下迁移状态。
// retired 允许重新发布，用于保留期内的原子回滚；failed 必须新建版本重试。
func CanTransitionIndexVersion(from, to IndexVersionState) bool {
	switch from {
	case IndexStaging:
		return to == IndexPublished || to == IndexFailed
	case IndexPublished:
		return to == IndexRetired
	case IndexRetired:
		return to == IndexPublished
	default:
		return false
	}
}

// BuildTaskState 标识文档构建任务的可恢复执行状态。
type BuildTaskState string

const (
	// BuildQueued 表示任务等待 Worker 领取。
	BuildQueued BuildTaskState = "queued"
	// BuildRunning 表示任务持有有效租约并正在执行。
	BuildRunning BuildTaskState = "running"
	// BuildSucceeded 表示任务产物已安全落定。
	BuildSucceeded BuildTaskState = "succeeded"
	// BuildFailed 表示任务本次尝试已失败，可由显式重试重新排队。
	BuildFailed BuildTaskState = "failed"
	// BuildCancelled 表示任务已取消且不得继续执行。
	BuildCancelled BuildTaskState = "cancelled"
)

// CanTransitionBuildTask 判断任务状态迁移是否符合领取、恢复和显式重试语义。
func CanTransitionBuildTask(from, to BuildTaskState) bool {
	switch from {
	case BuildQueued:
		return to == BuildRunning || to == BuildCancelled
	case BuildRunning:
		return to == BuildQueued || to == BuildSucceeded || to == BuildFailed || to == BuildCancelled
	case BuildFailed:
		return to == BuildQueued || to == BuildCancelled
	default:
		return false
	}
}
