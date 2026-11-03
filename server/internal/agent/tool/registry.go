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

var codePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_.-]{0,119}$`)

type Registry struct {
	mu    sync.RWMutex
	tools map[string]Tool
}

func NewRegistry() *Registry {
	return &Registry{tools: make(map[string]Tool)}
}

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
	if definition.Parameters == nil {
		return fmt.Errorf("工具 %s 的参数定义不能为空", definition.Code)
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.tools[definition.Code]; exists {
		return fmt.Errorf("工具编码 %s 已注册", definition.Code)
	}
	r.tools[definition.Code] = implementation
	return nil
}

func (r *Registry) Get(code string) (Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	implementation, exists := r.tools[code]
	return implementation, exists
}

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
