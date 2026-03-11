package source

import "errors"

// ErrMissingCUEModule indicates the fetched artifact does not contain a CUE module.
var ErrMissingCUEModule = errors.New("artifact does not contain a cue module")
