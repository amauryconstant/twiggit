package cmdutil

import (
	"context"

	"twiggit/internal/core"
)

// HookRunner runs a hook against the request payload.
//
// This is the consumer-side interface declared in cmdutil; concrete
// implementations live in the infrastructure layer and are wired into
// the command factory. Declaring the contract here keeps the dependency
// direction one-way: cmd/ depends on cmdutil, not on infrastructure.
type HookRunner interface {
	Run(ctx context.Context, req *core.HookRunRequest) (*core.HookResult, error)
}
