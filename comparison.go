package gg

import "github.com/rschoonheim/gg/internal/grouping"

// ErrUniverseMismatch is returned when attempting to compare or combine
// two groupings defined over universes of different sizes. It is
// re-exported from the internal grouping package.
var ErrUniverseMismatch = grouping.ErrUniverseMismatch
