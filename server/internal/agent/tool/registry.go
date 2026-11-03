package agenttool

import (
	"errors"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"
	"sync"
)

// codePattern 限制工具编码为可持久化且适合模型调用的稳定格式。
var codePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_.-]{0,119}$`)

// Registry 并发安全地保存按稳定编码索引的工具。
type Registry struct {
	mu    sync.RWMutex    // mu 保护 tools。
	tools map[string]Tool // tools 按 Definition.Code 保存实现。
}

// NewRegistry 创建空工具注册表。
func NewRegistry() *Registry {
	return &Registry{tools: make(map[string]Tool)}
}

// Register 校验定义完整性和编码唯一性后注册工具。
func (r *Registry) Register(implementation Tool) error {
	if implementation == nil {
		return errors.New("工具不能为空")
	}
	definition := implementation.Definition()
	if !codePattern.MatchString(definition.Code) {
		return fmt.Errorf("工具编码 %q 无效", definition.Code)
	}
	if strings.TrimSpace(definition.Name) == "" {
		return fmt.Errorf("工具 %s 的名称不能为空", definition.Code)
	}
	if strings.TrimSpace(definition.Description) == "" {
		return fmt.Errorf("工具 %s 的描述不能为空", definition.Code)
	}
	if strings.TrimSpace(definition.Source) == "" {
		return fmt.Errorf("工具 %s 的来源不能为空", definition.Code)
	}
	if definition.Parameters == nil {
		return fmt.Errorf("工具 %s 的参数定义不能为空", definition.Code)
	}
	if definition.Timeout <= 0 {
		return fmt.Errorf("工具 %s 的超时必须大于零", definition.Code)
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.tools[definition.Code]; exists {
		return fmt.Errorf("工具编码 %s 已注册", definition.Code)
	}
	r.tools[definition.Code] = implementation
	return nil
}

// Get 按稳定编码返回工具实现。
func (r *Registry) Get(code string) (Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	implementation, exists := r.tools[code]
	return implementation, exists
}

// List 按稳定编码排序返回全部工具。
func (r *Registry) List() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	codes := slices.Sorted(maps.Keys(r.tools))
	tools := make([]Tool, 0, len(codes))
	for _, code := range codes {
		tools = append(tools, r.tools[code])
	}
	return tools
}
