package source

import "errors"

// ErrSourceNotFound indicates the referenced OCIRepository does not exist.
var ErrSourceNotFound = errors.New("source not found")

// ErrSourceNotReady indicates the OCIRepository exists but is not ready.
var ErrSourceNotReady = errors.New("source not ready")

// ErrMissingCUEModule indicates the fetched artifact does not contain a CUE module.
var ErrMissingCUEModule = errors.New("artifact does not contain a cue module")
