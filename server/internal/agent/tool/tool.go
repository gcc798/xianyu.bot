package agenttool

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/eino-contrib/jsonschema"
)

type BudgetClass string

const (
	BudgetGeneral BudgetClass = "general"
	BudgetSearch  BudgetClass = "search"
)

type Definition struct {
	Code             string
	Name             string
	Description      string
	Source           string
	ProviderCode     string
	Parameters       *jsonschema.Schema
	ReadOnly         bool
	RequiresApproval bool
	Timeout          time.Duration
	BudgetClass      BudgetClass
}

type ExecutionError struct {
	Code    string
	Message string
	Cause   error
}

func (e *ExecutionError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *ExecutionError) Unwrap() error { return e.Cause }

type Scope struct {
	OwnerID uint
}

type Tool interface {
	Definition() Definition
	Invoke(context.Context, Scope, json.RawMessage) (string, error)
}
