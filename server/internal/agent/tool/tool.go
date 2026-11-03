package agenttool

import (
	"context"
	"encoding/json"
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
	Parameters       *jsonschema.Schema
	ReadOnly         bool
	RequiresApproval bool
	Timeout          time.Duration
	BudgetClass      BudgetClass
}

type Scope struct {
	OwnerID uint
}

type Tool interface {
	Definition() Definition
	Invoke(context.Context, Scope, json.RawMessage) (string, error)
}
