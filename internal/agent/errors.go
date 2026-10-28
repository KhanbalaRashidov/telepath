package agent

import "errors"

// ErrResumeUnsupported is returned by Resume on adapters without session resume.
var ErrResumeUnsupported = errors.New("agent does not support resuming sessions")

// ErrCLINotFound is returned when an agent's CLI binary cannot be located.
var ErrCLINotFound = errors.New("agent CLI not found on PATH")
