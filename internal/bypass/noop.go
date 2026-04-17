package bypass

import "context"

// Noop is a no-op Bypasser used when no bypass service is configured.
// It satisfies the Bypasser interface but never performs any network call.
type Noop struct{}

// compile-time interface check
var _ Bypasser = Noop{}

// Enabled always returns false.
func (Noop) Enabled() bool { return false }

// Solve always returns ErrDisabled.
func (Noop) Solve(_ context.Context, _ string) (*Result, error) {
	return nil, ErrDisabled
}
