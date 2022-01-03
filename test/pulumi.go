package test

import (
	"context"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
)

// GetStack simple test function to retrieve the current stack from
// the machines local state
// Note: Only works for stack state persisted locally
func GetStack(ctx context.Context) (*auto.Stack, error) {
	s, err := auto.SelectStackLocalSource(ctx, "web", ".")
	if err != nil {
		return nil, err
	}
	return &s, nil
}
